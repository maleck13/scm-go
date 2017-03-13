package routes_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/maleck13/scm-go/config"
	"github.com/maleck13/scm-go/test"
	"github.com/stretchr/testify/assert"
)

const GIT_ROUTE_APP = "anothertestapp"

func GetTestPostData(repo string, clean string) map[string][]string {
	var vals = url.Values{"appGuid": {GIT_ROUTE_APP}, "repoKey": {test.TestGitPrivKey}, "repoUrl": {repo}, "repoBranch": {"master"},
		"canUpdate": {"true"}, "repoCommitHash": {""}, "repoType": {"branch"}, "clusterHost": {"testing.feedhenry.me"}, "clusterName": {"development"},
		"forceCleanClone": {clean}, "returnLogs": {"false"}, "token": {"ABbdy128763GG"}, "cacheKey": {"4cf888a75922f531a53dbc5aa18bb3fc"}, "fileKey": {"41dd75cfd308d11becedce623ee3ba9b"},
		"async": {"true"}, "repoId": {"development-kjwzwi4rlqnnnf5f6s44z76q"}}
	return vals
}

func CheckCommitPostData(appGuid, commit, repoType string) string {
	return `
	{
	  "repoUrl":"not important",
	  "repoBranch":"refs/heads/master",
	  "repoType":"` + repoType + `",
	  "bare":false,
	  "commit":"` + commit + `",
	  "repoCommitHash":"` + commit + `",
	  "clusterName":"development",
	  "appGuid":"` + appGuid + `"	
	}
	`
}

func TestTriggerBadRequest(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	var triggerUrl = fmt.Sprintf("%s/fhgithub/trigger", server.URL)

	var testBadPostData = url.Values{"canUpdate": {"true"}}

	response, err := http.DefaultClient.PostForm(triggerUrl, testBadPostData)

	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusBadRequest, response.StatusCode, "expected 400 status code")
}

func TestTriggerPullRequest(t *testing.T) {

	var (
		conf                    = config.Conf
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	//setup 2 repos
	repo1Params := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo2Params := test.CreateCommonGitParams(GIT_ROUTE_APP)
	testRepo := test.SetUpRepo(repo1Params)
	defer test.TearDown(testRepo.Workdir(), conf.GetRepoPath(repo2Params.RepoId()))

	//clone it once.
	chanOne := make(chan int)
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "true")
	<-chanOne

	//resend request will cause a pull
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "false")
	<-chanOne

}

func TestTriggerCleanCloneOnExistingRepo(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	//setup 2 repos
	repo1Params := test.CreateCommonGitParams(GIT_ROUTE_APP)
	testRepo := test.SetUpRepo(repo1Params)
	defer testRepo.Free()
	repo2Params := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo2 := test.SetUpRepo(repo2Params)
	defer testRepo2.Free()
	defer tearDown(testRepo.Workdir(), testRepo2.Workdir())
	var triggerUrl = fmt.Sprintf("%s/fhgithub/trigger", server.URL)
	response, err := http.PostForm(triggerUrl, GetTestPostData(testRepo2.Path(), "true"))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusOK, response.StatusCode, "expected 202 status code")
}

func TestTriggerCleanCloneConcurrent(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	//setup 2 repos
	repo1Params := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(repo1Params)
	defer tearDown(testRepo.Workdir())
	//clone it once.
	chanOne := make(chan int)
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "true")
	<-chanOne

	//clone it 4 times concurrently each clone removes the repo so if trampling it should cause a problem
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "true")
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "true")
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "true")
	go triggerAndAssert(t, testRepo.Path(), chanOne, server.URL, "true")

	for i := 0; i < 4; i++ {
		<-chanOne
	}

}

func triggerAndAssert(t *testing.T, repoPath string, done chan int, url string, clean string) {
	var vals = GetTestPostData(repoPath, clean)
	fmt.Println("post vals ", vals)
	var triggerUrl = fmt.Sprintf("%s/fhgithub/trigger", url)
	response, err := http.PostForm(triggerUrl, vals)
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusOK, response.StatusCode, "expected 200 status code")
	done <- 1
}

func TestDeleteRepo(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	repoParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(repoParams)
	defer tearDown(testRepo.Workdir())
	url := fmt.Sprintf("%s/fhgithub/delete_app", server.URL)
	reqData := test.GetGeneralPostData(repoParams.AppGuid)
	response, err := http.Post(url, "application/json", strings.NewReader(reqData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusOK, response.StatusCode, "expected 200 status code")
	_, err = os.Stat(testRepo.Workdir())
	assert.Error(t, err, "expected dir to be gone")

}

func TestDeleteNonExistRepo(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/delete_app", server.URL)
	reqData := test.GetGeneralPostData("notthere")
	response, err := http.Post(url, "application/json", strings.NewReader(reqData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusOK, response.StatusCode, "expected 200 status code") //no repo to delete no need to worry
}

func TestCheckCommitOk(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/check_commit", server.URL)
	repoParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(repoParams)
	defer tearDown(testRepo.Workdir())
	ref, err := testRepo.Head()
	assert.NoError(t, err, "did not expect an error")
	commit := ref.Target().String()
	reqData := CheckCommitPostData(repoParams.AppGuid, commit, "commit")
	response, err := http.Post(url, "application/json", strings.NewReader(reqData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusOK, response.StatusCode, "expected 200 status code")
	content, _ := ioutil.ReadAll(response.Body)
	resMap := make(map[string]string)
	err = json.Unmarshal(content, &resMap)
	assert.NoError(t, err, "unexpected error")
	assert.EqualValues(t, commit, resMap["value"], "expected commit to be same")
}

func TestCheckRef(t *testing.T) {

}

func TestCheckCommitNoRepo(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/check_commit", server.URL)
	reqData := CheckCommitPostData(test.TEST_APP_GUID, "doesntmatter", "branch")
	response, err := http.Post(url, "application/json", strings.NewReader(reqData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusNoContent, response.StatusCode, "expected 204 status code")
}

func TestCheckBadCommit(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/check_commit", server.URL)
	repoParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(repoParams)
	defer tearDown(testRepo.Workdir())
	reqData := CheckCommitPostData(repoParams.AppGuid, "doesntmatter", "commit")
	response, err := http.Post(url, "application/json", strings.NewReader(reqData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "expected 404 status code")

}

func TestCheckNoCommit(t *testing.T) { //back compat fh-scm sends a 404
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/check_commit", server.URL)
	repoParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(repoParams)
	defer tearDown(testRepo.Workdir())
	reqData := CheckCommitPostData(repoParams.AppGuid, "", "commit")
	response, err := http.Post(url, "application/json", strings.NewReader(reqData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	assert.Equal(t, http.StatusNotFound, response.StatusCode, "expected 404 status code")
}

func TestListRemotes(t *testing.T) {
	var (
		server *httptest.Server = test.SetUpIntegrationServer()
	)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/list_remote", server.URL)
	repoParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(repoParams)
	defer tearDown(testRepo.Workdir())
	postData := test.GetGeneralPostData(repoParams.AppGuid)
	response, err := http.Post(url, "application/json", strings.NewReader(postData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	content, err := ioutil.ReadAll(response.Body)
	resMap := make(map[string][]interface{})
	json.Unmarshal(content, &resMap)

	if _, ok := resMap["tags"]; !ok {
		t.Fatal("expected tags in map")
	}
	if _, ok := resMap["branches"]; !ok {
		t.Fatal("expected branches in map")
	}

}

func tearDown(paths ...string) {

	for _, p := range paths {
		log.Printf("removing path %s ", p)
		if err := os.RemoveAll(p); err != nil {
			log.Panic("error with test tear down " + err.Error())
		}
	}
}
