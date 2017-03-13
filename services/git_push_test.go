package services_test

import (
	"github.com/maleck13/scm-go/config"
	"github.com/maleck13/scm-go/services"
	"github.com/maleck13/scm-go/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v23"
	"testing"
)

var calledPush = false

var testPush = func(t *testing.T) services.Pusher {
	return func(refs []string, opts *git.PushOptions) error {
		assert.NotNil(t, refs, "expected refs")
		assert.NotNil(t, opts, "expected opts")
		calledPush = true
		return nil
	}
}

func TestPushToOrigin(t *testing.T) {
	calledPush = false
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.RepoKey = test.TestGitPrivKey
	testRepo := test.SetUpRepo(gitParams)
	assert.NotNil(t, testRepo)
	defer test.TearDown(testRepo.Workdir())
	testRepo.Remotes.Create("origin", "git://foo/bar")
	publicPrivateKey, err := services.SetUpSSHKeys(gitParams.RepoContext, gitParams.AppContext, conf.GetKeysPath)
	assert.NoError(t, err, "did not expect an error")
	err = services.PushToOrigin(gitParams.RepoContext, conf, gitParams, publicPrivateKey, testPush(t))
	assert.NoError(t, err, "did not expect an error")
	assert.True(t, calledPush, "expected push to be called")
}
