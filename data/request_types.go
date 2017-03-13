package data

import (
	"errors"
	"fmt"
	"gopkg.in/libgit2/git2go.v23"
	"net/url"
	"strings"
)

//TODO document the types here well

const (
	default_repo_branch = "master"
	commit_key          = "commit"
)

type REQUIRED uint64

const (
	REQUIRE_APP_GUID          REQUIRED = 0
	REQUIRE_CLUSTER_NAME      REQUIRED = 1
	REQUIRE_REPO_URL          REQUIRED = 2
	REQUIRE_FULL_FILE_PATH    REQUIRED = 3
	REQUIRE_COMMON_SCM_PARAMS REQUIRED = 4
	REQUIRE_COMMIT            REQUIRED = 5
	REQUIRE_REPO_COMMIT_HASH  REQUIRED = 6
)

//this is the root type created by all post requests with json or form data. To create a new instance you should use the constructor
// NewScmParams() *ScmRequestContext  or if it is form data then use NewScmParamsFromFormData(data url.Values) *ScmRequestContext
type ScmRequestContext struct {
	*AppContext
	*RepoContext
	*RepoPrepareActions
	*RequestContext
	*FileContext
}

func NewScmParams() *ScmRequestContext {
	return &ScmRequestContext{
		RepoPrepareActions: &RepoPrepareActions{Gitpullbeforepush: true},
		RepoContext:        &RepoContext{RepoBranch: default_repo_branch, RepoType: "branch", Bare: false, Local: false, Ref: &RefContext{}, BranchType: git.BranchRemote},
		AppContext:         &AppContext{},
		FileContext:        &FileContext{},
		RequestContext:     &RequestContext{},
	}
}

func NewScmParamsFromFormData(data url.Values) *ScmRequestContext {
	gParams := NewScmParams()
	gParams.AppContext.AppGuid = data.Get("appGuid")
	gParams.RepoContext.RepoKey = data.Get("repoKey")
	gParams.RepoContext.RepoUrl = data.Get("repoUrl")
	gParams.RepoContext.RepoBranch = data.Get("repoBranch")
	gParams.RepoContext.RepoCommitHash = data.Get("repoHash")
	gParams.RepoContext.RepoType = data.Get("repoType")
	gParams.RequestContext.ClusterHost = data.Get("clusterHost")
	gParams.RequestContext.ClusterName = data.Get("clusterName")
	gParams.RepoPrepareActions.ForceCleanClone = data.Get("forceCleanClone")
	gParams.RequestContext.ReturnLogs = data.Get("returnLogs")
	gParams.RequestContext.Token = data.Get("token")
	gParams.RequestContext.CacheKey = data.Get("cacheKey")
	gParams.RequestContext.FileKey = data.Get("fileKey")

	if "false" == data.Get("gitpullbeforepush") {
		gParams.RepoPrepareActions.Gitpullbeforepush = false
	}

	return gParams
}

func (gr *ScmRequestContext) ValidateParams(params ...REQUIRED) error {
	var errMsg string = "missing %s param"
	if nil == params {
		return nil
	}

	for _, r := range params {
		if r == REQUIRE_COMMON_SCM_PARAMS {
			if err := gr.ValidateCommonParams(); err != nil {
				return err
			}
		}
		if r == REQUIRE_APP_GUID && "" == gr.AppContext.AppGuid {
			return errors.New(fmt.Sprintf(errMsg, "appGuid"))
		} else if r == REQUIRE_CLUSTER_NAME && "" == gr.RequestContext.ClusterName {
			return errors.New(fmt.Sprintf(errMsg, "clusterName"))
		} else if r == REQUIRE_REPO_URL && "" == gr.RepoContext.RepoUrl {
			return errors.New(fmt.Sprintf(errMsg, "repoUrl"))
		} else if r == REQUIRE_FULL_FILE_PATH && "" == gr.FileContext.FullFilePath {
			return errors.New(fmt.Sprintf(errMsg, "fullfilepath"))
		} else if r == REQUIRE_COMMIT && "" == gr.Commit {
			return errors.New(fmt.Sprintf(errMsg, commit_key))
		} else if r == REQUIRE_REPO_COMMIT_HASH && "" == gr.RepoCommitHash {
			return errors.New(fmt.Sprintf(errMsg, "repoCommitHash"))
		}

	}
	return nil
}

func (gr *ScmRequestContext) ValidateCommonParams() error {
	var errMsg string = "missing %s param"
	if "" == gr.AppContext.AppGuid {
		return errors.New(fmt.Sprintf(errMsg, "appGuid"))
	} else if "" == gr.RequestContext.ClusterName {
		return errors.New(fmt.Sprintf(errMsg, "clusterName"))
	} else if "" == gr.RepoContext.RepoUrl {
		return errors.New(fmt.Sprintf(errMsg, "repoUrl"))
	}
	return nil
}

func (gr *ScmRequestContext) hasFile() bool {
	return gr.FileContext.Name != ""
}

func (gr *ScmRequestContext) Repo_Branch() string {
	if "" == gr.RepoContext.RepoBranch {
		return default_repo_branch
	}
	return gr.RepoContext.RepoBranch
}

func (gr *ScmRequestContext) CanUpdate() bool {
	return gr.RepoContext.RepoType == "branch"
}

func (rc *ScmRequestContext) RepoId() string {
	return rc.ClusterName + "-" + rc.AppGuid
}

//holds the values sent in the ref body param (as these are sent as both a json object and a string we have to do a little bit of fiddling in routes/middleware.go to get it be created correctly
type RefContext struct {
	Hash  string
	Type  string
	Value string
}

//encapsulates the parts of the json body that are specific to a repo
type RepoContext struct {
	RepoKey        string         `json:"repoKey"`
	RepoUrl        string         `json:"repoUrl"`
	RepoBranch     string         `json:"repoBranch"`
	RepoType       string         `json:"repoType"`
	Bare           bool           `json:"bare"`
	RepoCommitHash string         `json:"repoHash"`
	Commit         string         `json:"commit"`
	Local          bool           `json:"local"`
	Ref            *RefContext    `json:"-"`
	RefStr         interface{}    `json:"ref"` //todo this is gross equiv to using Object and shouldn't be needed. This is because sometimes we send stringified json and sometimes we send actual json
	BranchType     git.BranchType `json:"-"`
}

//////////////************ simplify begin.This is too complex *********************** //////

func (rc *RepoContext) HasCommit() bool {
	return ((commit_key == rc.Ref.Type) || (commit_key == rc.RepoType))
}

//this is required for dealing with inconsistency in our existing api
//checks if the sent type is a commit if it is a commit it constructs a Remote with the correct information for use with checking commits etc
func (rc *RepoContext) GetCommitDetails() (Remotes, error) {
	var remote Remotes
	if rc.Ref.Type == commit_key {
		remote = Remotes{
			Type:  rc.Ref.Type,
			Hash:  rc.Ref.Hash,
			Value: rc.Ref.Value,
		}
	} else if rc.Commit != "" && rc.RepoType == commit_key {
		remote = Remotes{
			Type:  commit_key,
			Value: rc.Commit,
			Hash:  rc.Commit,
		}
	} else if rc.RepoCommitHash != "" {
		remote = Remotes{
			Type:  rc.RepoType,
			Value: rc.RepoBranch,
			Hash:  rc.RepoCommitHash,
		}
	} else {
		return remote, errors.New("no repoHash or commit present when looking for commit details")
	}

	return remote, nil
}

//this is awkward because of back compat. sometimes we send a ref as an object sometimes as a string sometimes as part of the top level object
//checks if it is a commit and delegates to GetCommitDetails otherwise it checks if ref has been populated and constructs a  remote based on that data.
// if the ref object hasn't been constructed it pulls the data from the toplevel object.
//essentially reproduces this logic from fh-scm
/*
var ref = {type: gitReq.repoType, value: gitReq.repoBranch, hash: gitReq.repoCommitHash}; //
	if(req.body.commit){
		ref = {type:'commit', value: req.body.commit, hash:req.body.commit};
	} else if(req.body.ref){
		ref = req.body.ref;
		if(typeof ref === 'string'){
		ref = JSON.parse(ref);
	}
}*/
func (rc *RepoContext) RefDetails() (Remotes, error) {
	//commit deal with it and return
	if rc.HasCommit() {
		return rc.GetCommitDetails()
	}
	if "" == rc.RepoType || "" == rc.RepoBranch {
		return Remotes{}, errors.New("missing value repoType || repoBranch || repoHash")
	}
	//default
	remote := Remotes{
		Type:  rc.RepoType,
		Hash:  rc.RepoCommitHash,
		Value: rc.RepoBranch,
	}
	//it has been sent as {"ref":{}}
	if nil != rc.Ref && "" != rc.Ref.Value {
		return Remotes{
			Type:  rc.Ref.Type,
			Hash:  rc.Ref.Hash,
			Value: rc.Ref.Value,
		}, nil
	}

	return remote, nil
}

//todo this is coded this way in scm also with hard coded remotes origin would be nice to support more than just origin as a remote and also local branches
// this could be improved if we used the full ref value every where but that causes probs in ngui at the moment
/*
from fh-scm
function getRefValue(ref){
  var refValue = 'HEAD';
  if(ref){
    if(ref.value && ref.value.indexOf('refs/') > -1){
      //the value contains the direct reference
      refValue = ref.value;
    } else {
      if(ref.type === 'branch'){
        //the branch name is retrived via git ls-remote, see above
        //we have to specify branch this way because the remote branch may not be checked out locally
        refValue = 'refs/remotes/origin/' + ref.value;
      } else if(ref.type === 'tag'){
        refValue = 'refs/tags/' + ref.value;
      } else if(ref.type === 'commit'){
        refValue = ref.value;
      }
    }
  }
  return refValue;
}
*/
func (gr *RepoContext) GetRefValue() (string, error) {
	remoteInfo, err := gr.RefDetails()
	if nil != err {
		return "", err
	}

	var refVal = "HEAD"
	// full ref has been sent no need to alter
	if strings.Contains(remoteInfo.Value, "refs/") {
		refVal = remoteInfo.Value
	} else if "branch" == remoteInfo.Type {
		//default type is BranchRemote
		if gr.BranchType == git.BranchLocal {
			refVal = "refs/heads/" + remoteInfo.Value
		} else {
			refVal = "refs/remotes/origin/" + remoteInfo.Value
		}
	} else if "tag" == remoteInfo.Type {
		refVal = "refs/tags/" + remoteInfo.Value
	} else if commit_key == remoteInfo.Type {
		refVal = remoteInfo.Value
	}
	return refVal, err
}

///////////////************ simplify end *********************** //////

func (gr *RepoContext) GetRepoType() string {
	return gr.RepoType
}

//encapsulates the actions to be performed on a repo prior to doing the main body of work
type RepoPrepareActions struct {
	Gitpullbeforepush bool   `json:"gitpullbeforepush"`
	ForceCleanClone   string `json:"forceCleanClone"`
}

//encapsulates the parts of the request that are app specific
type AppContext struct {
	AppGuid string `json:"appGuid"`
}

//encapsulates the parts of the request that are context around the request (who asked for it auth tokens etc)
type RequestContext struct {
	ClusterHost string `json:"clusterHost"`
	ClusterName string `json:"clusterName"`
	ReturnLogs  string `json:"returnLogs"`
	Token       string `json:"token"`
	CacheKey    string `json:"cacheKey"`
	FileKey     string `json:"fileKey"`
}

type Remotes struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Hash  string `json:"hash"`
}

//encapsulates the parts of a request around dealing with a file(s)
type RequestFile struct {
	AuthorEmail string `json:"authorEmail"`
	AuthorName  string `json:"authorName"`
	Contents    string `json:"contents"`
	IsBinary    bool   `json:"isBinary"`
	IsDirectory bool   `json:"isDirectory"`
	Name        string `json:"name"`
	Path        string `json:"path"`
}

//composed with a RequestFile this adds additional functionality and data. Could perhaps merge the two together
type FileContext struct {
	*RequestFile `json:"file"`
	FullFilePath string   `json:"fullpath"`
	Files        []string `json:"files"`
}

func (fc *FileContext) isValid() error {
	if fc.FullFilePath == "" {
		return errors.New("expected full file path")
	}
	return nil
}

func (rf *RequestFile) IsValid() error {
	if rf.Name == "" {
		return errors.New("expected a file name")
	}
	if rf.AuthorEmail == "" {
		return errors.New("expected an author email")
	}
	if rf.AuthorName == "" {
		return errors.New("expected an author name")
	}
	if rf.Path == "" {
		return errors.New("expected a file path")
	}
	return nil

}

//composed of a repocontext it build on the data to help with checkout specific actions it implements the data/Checkout interface
type checkoutContext struct {
	*RepoContext
	BranchType git.BranchType
}

func (cc *checkoutContext) IsBranch() bool {
	return "branch" == cc.RepoType
}

func (cc *checkoutContext) IsTag() bool {
	return "tag" == cc.RepoType
}

func (cc *checkoutContext) IsCommit() bool {
	return commit_key == cc.RepoType
}

func (cc *checkoutContext) IsLocalBranch() bool {
	return cc.BranchType == git.BranchLocal
}

func (cc *checkoutContext) SetBranchType(branchType git.BranchType) {
	cc.BranchType = branchType
}

func (cc *checkoutContext) GetBranchType() git.BranchType {
	return cc.BranchType
}

func NewCheckoutContext(context *RepoContext) *checkoutContext {
	return &checkoutContext{
		context,
		git.BranchRemote,
	}
}

type TriggerResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewTriggerResponse() *TriggerResponse {
	return &TriggerResponse{Status: "queued", Message: "Update request has been queued."}
}
