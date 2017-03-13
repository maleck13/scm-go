package services_test

import (
	"github.com/maleck13/scm-go/config"
	"github.com/maleck13/scm-go/data"
	"github.com/maleck13/scm-go/services"
	"github.com/maleck13/scm-go/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPullRepoBadParams(t *testing.T) {

	conf := config.NewConfig(test.TEST_CONF)
	req := data.NewScmParams()
	req.RepoUrl = conf.GetRepoPath(test.TEST_GIT_REPO_NAME)
	req.ClusterName = "development"
	req.Bare = true
	err := services.PullRepo(req.RepoContext, conf, req, services.PublicPrivateKeyLocation{})
	assert.Error(t, err, "expected an error")
}

//todo flesh out further make changes to origin and assert changes made it after pull
func TestPullRepoOk(t *testing.T) {
	testRepo := test.SetUpRepo(test.CreateCommonGitParams(test.TEST_APP_GUID))
	defer testRepo.Free()
	defer test.TearDown(testRepo.Workdir())
	req := data.NewScmParams()
	req.AppGuid = test.TEST_GIT_CLONE_REPO
	req.RepoUrl = testRepo.Path()
	req.ClusterName = test.TEST_CLUSTER_NAME
	req.RepoKey = test.TestGitPrivKey

	conf := config.NewConfig(test.TEST_CONF)
	publicPrivateKey, err := services.SetUpSSHKeys(req.RepoContext, req.AppContext, conf.GetKeysPath)
	assert.NoError(t, err, "did not expect an error")
	path, err := services.CloneRepo(req.RepoContext, conf, req, publicPrivateKey)
	defer test.TearDown(conf.GetRepoPath(req.RepoId()))
	assert.NoError(t, err, "no error ecpected cloning repo")
	assert.NotEmpty(t, path, "did not expect empty path")
	err = services.PullRepo(req.RepoContext, conf, req, publicPrivateKey)
	assert.NoError(t, err, "did not expect an error pulling from remote")
}

//todo expand to make changes to remote repo and assert they where feched
func TestFetchOk(t *testing.T) {
	testRepo := test.SetUpRepo(test.CreateCommonGitParams(test.TEST_APP_GUID))
	defer test.TearDown(testRepo.Workdir())
	req := data.NewScmParams()
	req.AppGuid = test.TEST_GIT_CLONE_REPO
	req.RepoUrl = testRepo.Path()
	req.ClusterName = test.TEST_CLUSTER_NAME
	req.RepoKey = test.TestGitPrivKey

	conf := config.NewConfig(test.TEST_CONF)
	publicPrivateKey, err := services.SetUpSSHKeys(req.RepoContext, req.AppContext, conf.GetKeysPath)
	assert.NoError(t, err, "did not expect an error")
	path, err := services.CloneRepo(req.RepoContext, conf, req, publicPrivateKey)
	defer test.TearDown(conf.GetRepoPath(req.RepoId()))
	assert.NoError(t, err, "no error ecpected cloning repo")
	assert.NotEmpty(t, path, "did not expect empty path")

	err = services.FetchFromRemote(req.RepoContext, conf, req, "origin", publicPrivateKey)

	assert.NoError(t, err, "did not expect an error fetching from remote")
}

func TestFetchBadRepo(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.ClusterName = test.TEST_CLUSTER_NAME
	gitParams.RepoKey = test.TestGitPrivKey
	conf := config.NewConfig(test.TEST_CONF)
	publicPrivateKey, err := services.SetUpSSHKeys(gitParams.RepoContext, gitParams.AppContext, conf.GetKeysPath)
	assert.NoError(t, err, "did not expect an error")
	err = services.FetchFromRemote(gitParams.RepoContext, conf, gitParams, "origin", publicPrivateKey)
	assert.Error(t, err, "expected error with bad repo")
}

func TestFetchBadeRemote(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(gitParams)
	defer test.TearDown(testRepo.Workdir())
	conf := config.NewConfig(test.TEST_CONF)
	err := services.FetchFromRemote(gitParams.RepoContext, conf, gitParams, "notthere", services.PublicPrivateKeyLocation{})
	assert.Error(t, err, "expected an error with bad remote")

}

func TestMergeBadRepo(t *testing.T) {

}
