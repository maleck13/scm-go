package routes_test

import (
	"bufio"
	"fmt"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/test"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestArchiveOk(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	url := fmt.Sprintf("%s/fhgithub/archive", server.URL)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	postData := test.GetGeneralPostData(gitParams.AppGuid)
	response, err := http.Post(url, "application/json", strings.NewReader(postData))
	assert.NoError(t, err, "did not expect an err")
	assert.NotNil(t, response, "expected response not to be nil")
	fmt.Println(response.ContentLength, response.StatusCode)
	reader := bufio.NewReader(response.Body)
	zipFile := "/tmp/testarchive.zip"
	f, err := os.Create(zipFile)
	assert.NoError(t, err, "did not expect an error")
	defer os.Remove(zipFile)
	reader.WriteTo(f)
	info, err := os.Stat(zipFile)
	assert.NoError(t, err, "did not expect an err")
	assert.True(t, info.Size() > 0, "expected a file bigger than 0")

}
