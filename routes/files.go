package routes

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/gorilla/mux"
	"github.com/fheng/scm-go/config"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/logger"
	"github.com/fheng/scm-go/services"
)

// list of file paths
type fileList struct {
	FileList []string `json:"filelist"`
}

// response from create, update, delete
type fileActionResult struct {
	Status string `json:"status"`
	Commit string `json:"commit"`
}

//List the files in the repo at their current HEAD, excluding .git and without taking the ref into account.
// It handles a GET request with path /fhgithub/listfiles/{repo}
func ListFiles(rw http.ResponseWriter, req *http.Request) {
	var (
		config     = config.GetConfig()
		params     = mux.Vars(req)
		repoName   = params["repo"]
		receiver   = make(chan services.DirList)
		files      = make([]string, 0)
		basePath   = config.GetRepoPath(repoName)
		logger     = logger.Logger
		identifier = data.NewRepoIdentity(repoName)
	)

	//as scm has an inconsistent http api we cant use our validation here
	if !services.RepoExists(config, identifier) {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	go services.RecurseReadDir(basePath, basePath, receiver)
	for dirRes := range receiver {
		if nil != dirRes.Error {
			HandleRouteError(req.URL, rw, data.NewErrorJSONUnexpectedError(dirRes.Error.Error()), logger)
			return
		}
		files = append(files, dirRes.FileList...)
	}

	sort.Strings(files)
	encoder := json.NewEncoder(rw)
	encoder.Encode(&fileList{files})
}

//this endpoint relates to //This endpoint relates to fhgithub/listfiles/{repo} POST
// it uses the sent ref to list the files. Not sure this is actually used in reality as when you checkout out a new branch in studio
// it cleans the repo re clones and checks out the given ref. So the the list GET list above is used
func ListFilesForRef(rw http.ResponseWriter, req *http.Request) {
	var (
		config     = config.GetConfig()
		params     = mux.Vars(req)
		reqBody    = data.NewScmParams()
		repoName   = params["repo"]
		logger     = logger.Logger
		identifier = data.NewRepoIdentity(repoName)
		decoder    = json.NewDecoder(req.Body)
	)
	reqBody, _, err := decodeAndValidate(decoder, config)
	if nil != err {
		HandleRouteError(req.URL, rw, err.(*data.ErrorJSON), logger)
		return
	}

	list, err := services.LsForRef(identifier, reqBody.RepoContext, config)
	sort.Strings(list)
	encoder := json.NewEncoder(rw)
	encoder.Encode(&fileList{list})

}

//this endpoint relates to /fhgithub/getfile
// gets a file based on the passed ref and returns it
func GetFile(wr http.ResponseWriter, req *http.Request) {
	var (
		decoder = json.NewDecoder(req.Body)
		logger  = logger.Logger
		conf    = config.Conf
	)

	params, validator, err := decodeAndValidate(decoder, conf)
	if nil != err {
		HandleRouteError(req.URL, wr, err.(*data.ErrorJSON), logger)
		return
	}

	if err := validator.ValidateParams(data.REQUIRE_FULL_FILE_PATH); err != nil {
		HandleRouteError(req.URL, wr, data.NewErrorJSONBadRequest(err.Error()), logger)
		return
	}

	content, err := services.ReadFile(params, params.RepoContext, params.FileContext, conf)

	if e, ok := err.(*data.ErrorJSON); ok {
		HandleRouteError(req.URL, wr, e, logger)
		return
	}
	if err != nil {
		HandleRouteError(req.URL, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}

	wr.Write(content)
}

// this endpoint relates to /fhgithub/createfile
// pulls the repo
// it creates the file in the repo
// adds commits and pushs the new file
func CreateUpdateFile(wr http.ResponseWriter, req *http.Request) {
	var (
		decoder = json.NewDecoder(req.Body)
		logger  = logger.Logger
		conf    = config.Conf
		encoder = json.NewEncoder(wr)
		url     = req.URL
	)

	params, _, err := decodeAndValidate(decoder, conf)
	if nil != err {
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest("invalid request missing params "+err.Error()), logger)
		return
	}

	fileContext := params.FileContext
	if err := fileContext.IsValid(); err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest("invalid request missing params "+err.Error()), logger)
		return
	}
	lockRepo(params)
	defer unlockRepo(params)

	publicPrivateKey, err := services.SetUpSSHKeys(params.RepoContext, params.AppContext, conf.GetKeysPath)
	defer services.RemoveKeys(publicPrivateKey)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to setup ssh keys "+err.Error()), logger)
		return
	}

	if params.RepoPrepareActions.Gitpullbeforepush {
		if err := services.PullRepo(params.RepoContext, conf, params, publicPrivateKey); err != nil {
			//todo if there is an error revert to previous commit. fh-scm does this but don't see the reason?
			HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to pull latest changes into repo "+err.Error()), logger)
			return
		}
	}
	//create file
	filePath, err := services.CreateUpdate(params, conf, fileContext.RequestFile)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to create file "+err.Error()), logger)
		return
	}

	commit, err := services.CommitFileChanges(params, conf, []string{filePath.RelativePath}, "created file "+filePath.RelativePath)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to update git repo "+err.Error()), logger)
		return
	}
	if err := services.PushToOrigin(params.RepoContext, conf, params, publicPrivateKey, nil); err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to push to origin "+err.Error()), logger)
		return
	}

	encoder.Encode(&fileActionResult{
		Status: "ok",
		Commit: commit.Hash,
	})
}

//Delete the file from the repo. To do this we pull the repo, delete the file commit the changes and then push to the remote
//resonds to POST fhgithub/deletefile
func DeleteFile(wr http.ResponseWriter, req *http.Request) {
	var (
		decoder = json.NewDecoder(req.Body)
		logger  = logger.Logger
		conf    = config.Conf
		encoder = json.NewEncoder(wr)
		url     = req.URL
	)

	params, _, err := decodeAndValidate(decoder, conf)
	if nil != err {
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest("invalid request missing params "+err.Error()), logger)
		return
	}

	fileContext := params.FileContext
	if err := params.FileContext.IsValid(); err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONBadRequest("invalid request missing params "+err.Error()), logger)
		return
	}
	lockRepo(params)
	defer unlockRepo(params)

	publicPrivateKey, err := services.SetUpSSHKeys(params.RepoContext, params.AppContext, conf.GetKeysPath)
	defer services.RemoveKeys(publicPrivateKey)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}

	if params.RepoPrepareActions.Gitpullbeforepush {
		if err := services.PullRepo(params.RepoContext, conf, params, publicPrivateKey); err != nil {
			//if there is an error revert to previous commit
			//version of git2go currently not supporting revert (latest version does)
			HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to pull latest changes into repo "+err.Error()), logger)
			return
		}
	}

	filePath, err := services.DeleteFile(params, conf, fileContext.RequestFile)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to Delete File "), logger)
		return
	}

	commit, err := services.CommitFileChanges(params, conf, []string{filePath.RelativePath}, "deleted file "+filePath.RelativePath)
	if err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError("failed to commit File changes "+err.Error()), logger)
		return
	}
	if err := services.PushToOrigin(params.RepoContext, conf, params, publicPrivateKey, nil); err != nil {
		HandleRouteError(url, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}

	encoder.Encode(&fileActionResult{
		Status: "ok",
		Commit: commit.Hash,
	})
}
