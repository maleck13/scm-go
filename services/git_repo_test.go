package services_test

import (
	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/fheng/scm-go/config"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/services"
	"github.com/fheng/scm-go/test"
	"os"
	"testing"
)

func TestRepoExistsShouldBeFalse(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestRepoExistsShouldBeFalse in short mode. Not integration during short")
	}
	conf := config.NewConfig(test.TEST_CONF)
	repoName := "test"
	exists := services.RepoExists(conf, data.NewRepoIdentity(repoName))

	if true == exists {
		t.Fatal("repo should not exist")
	}
}

func TestRepoExistsShouldBeTrue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestRepoExistsShouldBeTrue in short mode. Not integration during short")
	}
	gitParams := test.CreateCommonGitParams(test.TEST_GIT_REPO_NAME)
	testRepo := test.SetUpRepo(gitParams)
	defer test.TearDown(testRepo.Workdir())

	conf := config.NewConfig(test.TEST_CONF)
	exists := services.RepoExists(conf, gitParams)

	if false == exists {
		t.Fatal("repo should exist")
	}
}

func TestRemoveRepoOk(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_GIT_REPO_NAME)
	conf := config.NewConfig(test.TEST_CONF)
	testRepo := test.SetUpRepo(gitParams)
	defer test.TearDown(testRepo.Workdir())
	err := services.RemoveRepo(conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	_, err = os.Stat(conf.GetRepoPath(gitParams.RepoId()))
	assert.Error(t, err, "expected an error")
	t.Log(err)
}
