package routes

import (
	"encoding/json"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/logger"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/services"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
	"net/http"
)

// Git pull, Git clone and Git Checkout
//maps to POST /fhgithub/trigger trigger_handler. Would much prefer this to have less logic and be broken up into a clone handler and a pull handler and a separate setup repo keys handler.
func PullCloneCheckout(wr http.ResponseWriter, req *http.Request) {

	var (
		encoder   *json.Encoder = json.NewEncoder(wr)
		gitParams *data.ScmRequestContext
		err       error
		config    *config.Config = config.GetConfig()
		mBump     *services.Bump = services.NewBump()
		logger                   = logger.Logger
		url                      = req.URL
	)

	if err = req.ParseForm(); nil != err {
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest("could not parse form data "+err.Error()), logger)
		return
	}
	//yes for some reason millicore sends form data
	gitParams = data.NewScmParamsFromFormData(req.Form)
	//lock this repo so only once clone, pull action can happen at a time
	lockRepo(gitParams)
	defer unlockRepo(gitParams)
	//set up the bump of millicore
	mBump.BumpTemplate = config.Millicore.Api_bump_version
	mBump.Enabled = config.Millicore.Enabled
	mBump.Params = gitParams
	defer mBump.BumpMillicore()

	if err = gitParams.ValidateCommonParams(); err != nil {
		mBump.CommandError = err
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest(err.Error()), logger)
		return
	}
	//would prefer to have the pull as a separate api but this is for back compat.
	// There is too much logic in one place here. Refactor requires millicore changes to call new apis.
	// note force clean is always passed when checking out a branch or tag for some reason. Not sure we need to do that seems a bit brute force
	if "true" == gitParams.RepoPrepareActions.ForceCleanClone {
		logger.Info("Cleaning repo for force clean ")
		if err = services.RemoveRepo(config, gitParams); err != nil {
			mBump.CommandError = err
			HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
			return
		}
	}

	publicPrivateKey, err := services.SetUpSSHKeys(gitParams.RepoContext, gitParams.AppContext, config.GetKeysPath)
	defer services.RemoveKeys(publicPrivateKey)
	if err != nil {
		mBump.CommandError = err
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}

	if services.RepoExists(config, gitParams) {
		if gitParams.CanUpdate() {
			err = services.PullRepo(gitParams.RepoContext, config, gitParams, publicPrivateKey)
		} else {
			// not sure what to do here fh-scm just returns
		}
	} else {
		//clone will checkout the branch passed as in git clone somerepo --branch=somebranch
		_, err = services.CloneRepo(gitParams.RepoContext, config, gitParams, publicPrivateKey)
	}
	if err != nil {
		mBump.CommandError = err
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}
	commitHash, err := services.GetHeadCommitHash(config, gitParams)
	mBump.CommitHash = commitHash

	//note tags and commits are essentially the same thing
	if "tag" == gitParams.RepoType || "commit" == gitParams.RepoType {
		checkout := data.NewCheckoutContext(gitParams.RepoContext)
		commitHash, err = services.Checkout(config, gitParams, checkout)
		if nil != err {
			mBump.CommandError = err
			HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("checking out "+gitParams.RepoType+" "+err.Error()), logger)
			return
		}
	}

	wr.WriteHeader(http.StatusOK)
	encoder.Encode(data.NewTriggerResponse())
}

//This responds to POST fhgithub/list_remote and lists the branches and tags of for the requested repo
func ListRemotes(wr http.ResponseWriter, req *http.Request) {

	var (
		encoder *json.Encoder  = json.NewEncoder(wr)
		decoder *json.Decoder  = json.NewDecoder(req.Body)
		config  *config.Config = config.Conf
		logger                 = logger.Logger
		url                    = req.URL
	)

	params, _, err := decodeAndValidate(decoder, config)
	if nil != err {
		HandleRouteError(url, wr, err.(*data.ErrorJSON), logger)
		return
	}

	remotes, err := services.ListBranchesAndTags(git.BranchRemote, config, params)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}

	err = encoder.Encode(remotes)

	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}

}

//This responds to POST fhgithub/check_commit it checks that the passed git ref is valid for the requested repo
func CheckReference(wr http.ResponseWriter, req *http.Request) {
	var (
		encoder *json.Encoder  = json.NewEncoder(wr)
		decoder *json.Decoder  = json.NewDecoder(req.Body)
		config  *config.Config = config.Conf
		logger                 = logger.Logger
		url                    = req.URL
	)

	params, _, err := decodeAndValidate(decoder, config)
	if nil != err {
		HandleRouteError(url, wr, err.(*data.ErrorJSON), logger)
		return
	}
	lockRepo(params)
	defer unlockRepo(params)
	fullCommit, err := services.CheckCommitOrRef(params.RepoContext, config, params)

	if nil != err {
		HandleRouteError(url, wr, data.NewErrorJSONNotFound(err.Error()), logger)
		return
	}

	encoder.Encode(fullCommit)

}

//This responds to POST /fhgithub/delete_app. It deletes the repo from disk
func DeleteRepo(wr http.ResponseWriter, req *http.Request) {
	var (
		decoder *json.Decoder  = json.NewDecoder(req.Body)
		config  *config.Config = config.Conf
		logger                 = logger.Logger
		params                 = data.NewScmParams()
		url                    = req.URL
	)

	if err := decoder.Decode(params); err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest("failed to decode json"+err.Error()), logger)
		return
	}

	if err := services.RemoveRepo(config, params); err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to remove repo"+err.Error()), logger)
		return
	}
}
