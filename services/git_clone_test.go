package services_test

import (
	"testing"

	"strconv"
	"time"

	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/fheng/scm-go/config"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/services"
	"github.com/fheng/scm-go/test"
)

func TestCloneRepoBadParams(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	req := data.NewScmParams()
	req.RepoUrl = conf.GetRepoPath(test.TEST_GIT_REPO_NAME)
	req.ClusterName = "development"
	req.Bare = true
	_, err := services.CloneRepo(req.RepoContext, conf, req, services.PublicPrivateKeyLocation{})
	assert.Error(t, err, "expected an error")

}

func TestCloneRepoBadGitUrl(t *testing.T) {

	conf := config.NewConfig(test.TEST_CONF)
	req := data.NewScmParams()
	req.RepoUrl = conf.GetRepoPath(test.TEST_GIT_REPO_NAME)
	req.ClusterName = "development"
	req.Bare = true
	req.RepoKey = test.TestGitPrivKey
	publicPrivateKey, err := services.SetUpSSHKeys(req.RepoContext, req.AppContext, conf.GetKeysPath)
	assert.NoError(t, err, "did not expect an error")
	_, err = services.CloneRepo(req.RepoContext, conf, req, publicPrivateKey)
	assert.Error(t, err, "expected an error")

}

func TestCloneRepoSuccess(t *testing.T) {
	testRepo := test.SetUpRepo(test.CreateCommonGitParams(test.TEST_APP_GUID))
	defer test.TearDown(testRepo.Workdir())
	req := data.NewScmParams()
	timeStamp := time.Now().Nanosecond()
	req.AppGuid = test.TEST_GIT_CLONE_REPO + strconv.Itoa(timeStamp)
	req.RepoUrl = testRepo.Path()
	req.ClusterName = test.TEST_CLUSTER_NAME
	req.RepoKey = test.TestGitPrivKey

	conf := config.NewConfig(test.TEST_CONF)
	publicPrivateKey, err := services.SetUpSSHKeys(req.RepoContext, req.AppContext, conf.GetKeysPath)
	assert.NoError(t, err, "did not expect an error")
	path, err := services.CloneRepo(req.RepoContext, conf, req, publicPrivateKey)
	defer test.TearDown(conf.GetRepoPath(req.RepoId()))
	assert.NoError(t, err, "did not expect an error")
	assert.NotEmpty(t, path, "expected a git path")

}
