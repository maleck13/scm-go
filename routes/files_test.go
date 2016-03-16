package routes_test

import (
	"encoding/json"
	"fmt"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/test"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type FileData func() string

func GetFileTestPostData(repo, commitHash, file, appGuid string) string {
	return `
	{
	  "repoUrl":"` + repo + `",
	  "repoBranch":"refs/heads/master",
	  "repoType":"branch",
	  "bare":false,
	  "repoCommitHash":"` + commitHash + `",
	  "clusterName":"development",
	  "fullpath":"` + file + `",
	  "appGuid":"` + appGuid + `"	
	}
	
	`
}

func fileData(name, content string) FileData {
	return func() string {
		return `{
	    	"authorEmail":"test@test.com",
	    	"authorName":"test",
	    	"contents":"` + content + `",
	    	"isBinary":false,
	    	"isDirectory":false,
	    	"name":"` + name + `",
	    	"path":"/"
	  }`
	}
}

func badFileData() FileData {
	return func() string {
		return `{
	    	"authorEmail":"test@test.com",
	    	"authorName":"test",
	    	"contents":"",
	    	"isBinary":false,
	    	"isDirectory":false,
	    	"name":"",
	    	"path":"/"
	  }`
	}
}

func CreateFilePostData(repo, file, appGuid string, fileData FileData) string {
	key := strings.Replace(test.TestGitPrivKey, "\n", "\\n", -1)
	return `
	{
	  "repoKey":"` + key + `",	
	  "repoUrl":"` + repo + `",
	  "repoBranch":"refs/heads/master",
	  "repoType":"branch",
	  "bare":false,
	  "repoCommitHash":"",
	  "clusterName":"development",
	  "fullpath":"` + file + `",
	  "appGuid":"` + appGuid + `",
	  "local":true,
	  "file":` + fileData() + `
	}	
	`
}

func TestShouldListAllFilesInRepo(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()

	gitParams := test.CreateCommonGitParams(test.TEST_GIT_REPO_NAME)
	url := fmt.Sprintf("%s/fhgithub/listfiles/%s", server.URL, "development-"+gitParams.AppGuid)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err, "did not expect err")

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "should be 200")
	con, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err, "did not epext and error reading the body")
	var jMap map[string][]string
	json.Unmarshal(con, &jMap)
	k, v := jMap["filelist"]
	assert.True(t, v, "expected filelist to exist ")
	sort.Strings(k)
	ind := sort.SearchStrings(k, "/README")
	fmt.Printf("index is %d ", ind)
	assert.True(t, ind != len(k), "expected file to be in the list")
}

func TestShouldFailToListAllFilesInNonRepo(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/listfiles/%s", server.URL, "development-idontexist")
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "should be 204")
}

func TestShouldListFilesForRef(t *testing.T) {
	t.Skip("NOT YET IMPLIMENTED TestShouldListFilesForRef")
}

func TestShouldGetFileOk(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/getfile", server.URL)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	ref, err := repo.Head()
	assert.NoError(t, err, "did not expect error")

	commitHash := ref.Target().String()
	reqBody := GetFileTestPostData(repo.Path(), commitHash, "README", gitParams.AppGuid)
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 200, resp.StatusCode, "expected a 200 status code")
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err, "did not expect an error")
	fmt.Println(string(body))
	assert.True(t, resp.ContentLength > 0, "expected content")
}

func TestShouldGet404(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/getfile", server.URL)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	ref, err := repo.Head()
	assert.NoError(t, err, "did not expect error")
	commitHash := ref.Target().String()
	reqBody := GetFileTestPostData(repo.Path(), commitHash, "NOTTHERE", gitParams.AppGuid)
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 404, resp.StatusCode, "expected a 404 status code")
}

func TestShouldGetBadRequest(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/getfile", server.URL)
	reqBody := GetFileTestPostData("/tmp/blah", "hash", "README", "")
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 400, resp.StatusCode, "expected a 400 status code")
}

func TestShouldCreateNewFile(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	conf := config.NewConfig(test.TEST_CONF)
	url := fmt.Sprintf("%s/fhgithub/createfile", server.URL)
	timeStamp := time.Now().Nanosecond()
	appid := test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	params, path, repo := test.SetUpRepoAndCloneToLocalRepo(appid)
	defer test.TearDown(conf.GetRepoPath(params.RepoId()), repo)
	reqBody := CreateFilePostData(path, "test.txt", appid, fileData("test.txt", "test"))
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 200, resp.StatusCode, "expected a 200 status code")
}

func TestShouldUpdateExistingFile(t *testing.T) {
	server := test.SetUpIntegrationServer()
	conf := config.NewConfig(test.TEST_CONF)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/updatefile", server.URL)
	timeStamp := time.Now().Nanosecond()
	appid := test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	params, cloned, repo := test.SetUpRepoAndCloneToLocalRepo(appid)
	defer test.TearDown(conf.GetRepoPath(params.RepoId()), repo)
	reqBody := CreateFilePostData(cloned, "README", appid, fileData("README", "updated"))
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 200, resp.StatusCode, "expected a 200 status code")
	//read file check content
	file, err := os.Open(cloned + "/README")
	assert.NoError(t, err, "no error expected")
	content, err := ioutil.ReadAll(file)
	assert.NoError(t, err, "did not expect err")
	assert.EqualValues(t, content, "updated")
}

func TestUpdateShouldNotFailToUpdateNonExistingFile(t *testing.T) { //fh-scm back compat on an update if file is missing it just creates it
	server := test.SetUpIntegrationServer()
	conf := config.NewConfig(test.TEST_CONF)
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/updatefile", server.URL)
	timeStamp := time.Now().Nanosecond()
	appid := test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	params, cloned, repo := test.SetUpRepoAndCloneToLocalRepo(appid)
	defer test.TearDown(conf.GetRepoPath(params.RepoId()), repo)
	reqBody := CreateFilePostData(cloned, "NOTTHERE", appid, fileData("NOTHERE", "updated"))
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 200, resp.StatusCode, "expected a 404 status code")
	file, err := os.Open(cloned + "/NOTHERE")
	assert.NoError(t, err, "no error expected")
	content, err := ioutil.ReadAll(file)
	assert.NoError(t, err, "did not expect err")
	assert.EqualValues(t, content, "updated")
}

func TestShouldGetBadRequestMissingParams(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	conf := config.NewConfig(test.TEST_CONF)
	url := fmt.Sprintf("%s/fhgithub/updatefile", server.URL)
	timeStamp := time.Now().Nanosecond()
	appid := test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	params, cloned, repo := test.SetUpRepoAndCloneToLocalRepo(appid)
	defer test.TearDown(conf.GetRepoPath(params.RepoId()), repo)
	reqBody := CreateFilePostData(cloned, "NOTTHERE", appid, badFileData())
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 400, resp.StatusCode, "expected a 400 status code")
}

func TestShouldDeleteExistingFile(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	conf := config.NewConfig(test.TEST_CONF)
	url := fmt.Sprintf("%s/fhgithub/deletefile", server.URL)
	timeStamp := time.Now().Nanosecond()
	appid := test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	params, cloned, repo := test.SetUpRepoAndCloneToLocalRepo(appid)
	defer test.TearDown(conf.GetRepoPath(params.RepoId()), repo)
	reqBody := CreateFilePostData(cloned, "README", appid, fileData("README", "updated"))
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 200, resp.StatusCode, "expected a 200 status code")
	//check file gone
	_, err = os.Open(cloned + "/README")
	assert.Error(t, err, "expected an error opening deleted file")

}

func TestShouldNotFailDeletingNonFile(t *testing.T) { //back compat fh-scm does not error when file doesn't exist
	server := test.SetUpIntegrationServer()
	defer server.Close()
	conf := config.NewConfig(test.TEST_CONF)
	url := fmt.Sprintf("%s/fhgithub/deletefile", server.URL)
	timeStamp := time.Now().Nanosecond()
	appid := test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	params, cloned, repo := test.SetUpRepoAndCloneToLocalRepo(appid)
	defer test.TearDown(conf.GetRepoPath(params.RepoId()), repo)
	reqBody := CreateFilePostData(cloned, "NOTTHERE", appid, fileData("NOTTHERE", "updated"))
	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	assert.NoError(t, err, "did not expect err")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, resp, "did not expect a nil response")
	assert.EqualValues(t, 200, resp.StatusCode, "expected a 200 status code")

}
