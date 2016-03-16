package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"runtime"
	"time"

	"runtime/debug"
	"strconv"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/logger"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/routes"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/services"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
)

func init() {
	config.NewConfig(TEST_CONF)
	err := os.RemoveAll("/tmp/fh-scm/*")

	fmt.Println("REMOVED DIRECTORIES FOR TESTS", err)
	SetUpDirectories(config.Conf)
	fmt.Println("SETUP DIRECTORIES FOR TESTS")

	//could setup server here

}

const TEST_GIT_REPO_NAME = "testgit"
const TEST_GIT_BRANCH = "test"
const TEST_CONF = "../config/dev.json"
const TEST_APP_GUID = "testguid"
const TEST_APP_GUID2 = "testguid2"
const TEST_COMMIT_HASH_REPO = "testcommithash"
const TEST_GIT_CLONE_REPO = "testclone"
const TEST_GIT_CLONE_LOCATION = "/tmp/fh-scm/development-" + TEST_GIT_CLONE_REPO
const TEST_CLUSTER_NAME = "development"
const TEST_REPO_URL = "/tmp/fh-scm/testguid/.git/"
const TEST_TAG = "refs/tags/atag"

var TestGitPrivKey string = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0oebo4qY3Q96DsB11AjuTFYc2Hi44YW03xWj1IoUhqeqEpN8
GOkxIKv5C5apTaO0C83yJIB7n8oX7kJ6Ei1kUy0Ysyhov9belpUwfXiCJOuT7E1D
WHF7cpkrC39OY3n1dfBQOknOASXrppvfoJo0SAcnq/QLilyS4YmMOfu0d2oMa89r
J29m2phApH7ES0QwQW1v/mkbQDWLEUN+9b/RGeDS2QuWg0+rPkk9+KzimW5Uyiol
4jA67d63C4VoDfQLgY7RvtwtjTa6JpK4ag/jt7XazawfJLWwSUy1TSfMZHhiMaLy
1haqfu6a7VcvIZ3xWX3bubZZ0tMrpBWtmZYMXQIDAQABAoIBACm5wMoYBRYDJEDa
OkAtCPXON2db/eCMvH1eb5sHRuwtPNLzsivJw/fivbhjQujlYiDYz34WjOnthvKs
8D68Cl9F19hiUOh8sJ8NsI+dm8WvfyDda1STIVFkNBjSQXoLyO94Ep6K1C58Eolx
6U+OYIBKEr3M36CrmlzjAwzW6zyjRw6iXakXwFIWjMvbwiF1ozsmCaqz+ms0bxmN
j0JvyduIR6GnRGxNBYsJWjrjxRh9KWmvK+qy6AJ0WMRf0vAZuxalc8KJ7CDA0/Zv
yEuPcPY89gHKAN+u/Xr9ffJHT9q7/HdKsmfjjh0TOenu+xnj2BuSJgHDI6s1Zp5L
50UQNu0CgYEA6NKBSLWQXTQExqzEsmkP/TbNHpgrHyIKk9AYjXiaQs/u2DDbZ2pP
uBdpTMBk2pqS6XBDME5mHrwEwDs+dg9AuVZjcKwaYYwG4a71fSlV6dySIQfO/OBw
AcE1fKh66fFeGqya2GLRfg+feG4QdyvoyfVMu8VRgkeAVKcPRd1OeucCgYEA53z5
ndp6eWxq28Z7nL+ysp86CC+EbB7SNdSyOr9KAvq7a+LsSPbf8gwIf4OR13eHt5u3
+x46mFWiy1dPurZ0MHqB1aHj0jgn97NlhPo2gbxvOdjIwuu478nXJgrhi1rN+2Dc
yfeJuKvoVTmWXKaOs6HgP1pTbjFe04aLRaO7ehsCgYEAx9DScEqTScp5RuHENrKA
rHs588X5TFD07wMr707QZgL7I8RaqRjOrlo5X0vHwu3ipEJbB7xhXINVOH4gD0br
76S611p9kbaRuWNlATRWrc1GBN8kwFcHChp+Ayy9wMYMU6xLgObekOhrnvonW3/t
3FaQsp6YY81l0EFIlFdpIYUCgYAj7s1ciHZsdLzdoqC7gzI88eRzMtUOZ3Zf7sU/
w0O81KjVJWNiLBg1HVMQYB76YE6L3LshWt7jmJ9tiv8QT5QXllyjCb8weZESrSE8
FA7z8okdZJ49S5PofN9Nw3ChThYdSDrdivQw21Z0LD0/4obSMXV6wA2lVDqRARiL
VdAbMwKBgQCqdp8onVrA4Ug+WnjYiKsJ5605soPdOqmtP833PWF+0+Xm4meh7JbW
1ev7HE8eGxvVqVAV7pNKa0ddmlq15kSMbF9r9VQtWqXbR8gb0DhEfZKE45x97caQ
9TYXm8nZvQeeZpazOglpnLuFEg4wTUzm6lK5qqzLERX/C5Ib6D5Ubg==
-----END RSA PRIVATE KEY-----`

var TestPubKey string = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDSh5ujipjdD3oOwHXUCO5MVhzYeLjhhbTfFaPUihSGp6oSk3wY6TEgq/kLlqlNo7QLzfIkgHufyhfuQnoSLWRTLRizKGi/1t6WlTB9eIIk65PsTUNYcXtymSsLf05jefV18FA6Sc4BJeumm9+gmjRIByer9AuKXJLhiYw5+7R3agxrz2snb2bamECkfsRLRDBBbW/+aRtANYsRQ371v9EZ4NLZC5aDT6s+ST34rOKZblTKKiXiMDrt3rcLhWgN9AuBjtG+3C2NNromkrhqD+O3tdrNrB8ktbBJTLVNJ8xkeGIxovLWFqp+7prtVy8hnfFZfdu5tlnS0yukFa2Zlgxd`

func checkFatal(err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatalf("Unable to get caller")
	}
	log.Fatalf("Fail at %v:%v; %v", file, line, err)

	debug.PrintStack()
}

func UpdateReadMe(content, path string) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.FileMode(0666))
	checkFatal(err)
	defer file.Close()
	_, err = file.Write([]byte(content))
	checkFatal(err)

}

func CreateCommonGitParams(appId string) *data.ScmRequestContext {
	timeStamp := time.Now().Nanosecond()
	gitParams := data.NewScmParams()
	gitParams.AppGuid = fmt.Sprintf("%s%s", appId, strconv.Itoa(timeStamp))
	gitParams.ClusterName = TEST_CLUSTER_NAME
	gitParams.RepoUrl = TEST_REPO_URL
	gitParams.RepoContext.BranchType = git.BranchLocal
	return gitParams
}

func CreateCommit(repo *git.Repository, conf *config.Config) (*git.Commit, *git.Signature) {
	sig := &git.Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Now(),
	}

	idx, err := repo.Index()
	checkFatal(err)
	files := []string{"README", "test/"}
	err = idx.AddAll(files, git.IndexAddDefault, nil)
	checkFatal(err)
	treeId, err := idx.WriteTree()
	checkFatal(err)
	currentBranch, err := repo.Head()
	var currentTip *git.Commit

	if nil != currentBranch {
		currentTip, err = repo.LookupCommit(currentBranch.Target())
		checkFatal(err)
	}

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(err)
	var commitId *git.Oid
	if nil != currentTip {
		commitId, err = repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		checkFatal(err)
	} else {
		commitId, err = repo.CreateCommit("HEAD", sig, sig, message, tree) //initial commit no parent
		checkFatal(err)
	}
	checkFatal(err)
	commit, err := repo.LookupCommit(commitId)
	checkFatal(err)
	return commit, sig
}

func CreateReadMe(content, path string) {
	err := ioutil.WriteFile(path, []byte(content), os.ModePerm)
	checkFatal(err)
}

func createDirAndFile(rootPath string) {
	err := os.MkdirAll(rootPath+"/test", os.ModePerm)
	checkFatal(err)
	err = ioutil.WriteFile(rootPath+"/test/testfile.txt", []byte("test content"), os.ModePerm)
	checkFatal(err)
}

func createTag(repo *git.Repository, commit *git.Commit, tag string, sig *git.Signature) {
	_, err := repo.Tags.Create("atag", commit, sig, "tag")
	checkFatal(err)
}

func SetUpRepo(gitParams *data.ScmRequestContext) *git.Repository {
	var (
		err  error
		repo *git.Repository
	)
	conf := config.NewConfig(TEST_CONF)
	if repo, err = git.InitRepository(conf.GetRepoPath(gitParams.RepoId()), false); err != nil {
		log.Panic("failed to set up git test " + err.Error())
	}

	CreateReadMe("test", conf.GetRepoPath(gitParams.RepoId())+"/README")
	createDirAndFile(repo.Workdir())
	commit, sig := CreateCommit(repo, conf)
	createTag(repo, commit, "atag", sig)
	_, err = repo.CreateBranch(TEST_GIT_BRANCH, commit, false)

	checkFatal(err)

	return repo
}

func TearDown(paths ...string) {

	for _, p := range paths {
		log.Printf("removing path %s ", p)
		if err := os.RemoveAll(p); err != nil {
			log.Panic("error with test tear down " + err.Error())
		}
	}
}

func SetUpDirectories(conf *config.Config) {

	if _, err := os.Stat(conf.GetKeysPath().URL); err != nil {
		if err = os.MkdirAll(conf.GetKeysPath().URL, 0755); err != nil {
			log.Fatalln("failed to set up dirs " + err.Error())
		}
	}

	if _, err := os.Stat(conf.GetArchivePath()); err != nil {
		if err = os.MkdirAll(conf.GetArchivePath(), 0755); err != nil {
			log.Fatalln("failed to set up dirs " + err.Error())
		}
	}
}

func SetUpIntegrationServer() *httptest.Server {
	config := config.NewConfig(TEST_CONF)
	logger := logger.InitLogger(config.Logger)
	return httptest.NewServer(routes.SetUpRoutes(logger, config))
}

func SetUpRepoAndCloneToLocalRepo(cloneTo string) (*data.ScmRequestContext, string, string) {
	conf := config.NewConfig(TEST_CONF)
	repo1Params := CreateCommonGitParams(TEST_APP_GUID)
	repo2Params := CreateCommonGitParams(cloneTo)
	repo1 := SetUpRepo(repo1Params)
	repo2Params.AppGuid = cloneTo
	repo2Params.RepoUrl = repo1.Path()
	repo2Params.ClusterName = TEST_CLUSTER_NAME
	repo2Params.RepoKey = TestGitPrivKey
	publicPrivateKey, err := services.SetUpSSHKeys(repo2Params.RepoContext, repo2Params.AppContext, conf.GetKeysPath)
	checkFatal(err)
	path, err := services.CloneRepo(repo2Params.RepoContext, conf, repo2Params, publicPrivateKey)
	checkFatal(err)
	return repo2Params, path, repo1.Workdir()
}

func GetGeneralPostData(appGuid string) string {
	return `
	{
	  "repoUrl":"not important",
	  "repoBranch":"refs/heads/master",
	  "repoType":"branch",
	  "bare":false,
	  "commit":"doesntmatter",
	  "clusterName":"development",
	  "appGuid":"` + appGuid + `"	
	}
	`
}
