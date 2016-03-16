package services_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/services"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/test"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func init() {
	config.NewConfig(test.TEST_CONF)
}

type MockClient struct {
	ResponseStatus int
	error          error
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	resp := `{}`
	fmt.Println("mock client called ", m.error)
	return &http.Response{
		StatusCode: m.ResponseStatus,
		Body:       nopCloser{bytes.NewBufferString(resp)},
	}, m.error
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestBumpUrl(t *testing.T) {
	conf := config.GetConfig()
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	bump := services.Bump{}
	bump.BumpTemplate = conf.Millicore.Api_bump_version
	bump.Params = gitParams
	url := bump.BumpUrl()
	assert.NotEmpty(t, url, "expected a url ")

}

func TestBumpRequest(t *testing.T) {
	conf := config.GetConfig()
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.Token = "testtoken"
	gitParams.FileKey = "testkey"
	gitParams.CacheKey = "cachekey"
	bump := services.NewBump()
	bump.BumpClient = &MockClient{http.StatusOK, nil}
	bump.Params = gitParams
	bump.BumpTemplate = conf.Millicore.Api_bump_version
	bump.CommitHash = "testhash"

	br := services.NewBumpRequest(bump)

	assert.NotNil(t, br, "did not expect bump request to be nil")

	assert.Equal(t, br["token"], "testtoken")
	assert.Equal(t, br["cacheKey"], "cachekey")
	assert.Equal(t, br["fileKey"], "testkey")
}

func TestBumpRequestMockResponseOk(t *testing.T) {
	conf := config.GetConfig()
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.Token = "testtoken"
	gitParams.FileKey = "testkey"
	gitParams.CacheKey = "cachekey"
	bump := services.NewBump()
	bump.BumpClient = &MockClient{http.StatusOK, nil}
	bump.Params = gitParams
	bump.BumpTemplate = conf.Millicore.Api_bump_version
	bump.CommitHash = "testhash"

	err := bump.BumpMillicore()

	assert.NoError(t, err, "no error expected")
}

func TestBumpRequestMockResponseError(t *testing.T) {
	conf := config.GetConfig()
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.Token = "testtoken"
	gitParams.FileKey = "testkey"
	gitParams.CacheKey = "cachekey"

	bump := services.NewBump()
	bump.Enabled = true
	bump.BumpClient = &MockClient{http.StatusBadRequest, errors.New("connection refused")}
	bump.Params = gitParams
	bump.BumpTemplate = conf.Millicore.Api_bump_version
	bump.CommitHash = "testhash"

	err := bump.BumpMillicore()

	assert.Error(t, err, "error expected")
}
