package services_test

import (
	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/fheng/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
	"github.com/fheng/scm-go/config"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/services"
	"github.com/fheng/scm-go/test"
	"testing"
)

func TestCheckoutBranchOk(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.RepoContext.RepoBranch = test.TEST_GIT_BRANCH
	testRepo := test.SetUpRepo(gitParams)
	defer test.TearDown(testRepo.Workdir())

	branch, err := testRepo.LookupBranch(test.TEST_GIT_BRANCH, git.BranchLocal)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, branch, "expected a branch")

	checkout := data.NewCheckoutContext(gitParams.RepoContext)
	ref, err := services.CheckoutBranch(conf, gitParams, checkout)
	assert.NoError(t, err, "did not expect an error")
	assert.Equal(t, ref, branch.Reference.Name(), "should be same reference")
	gitParams.RepoBranch = "master"
	checkout = data.NewCheckoutContext(gitParams.RepoContext)
	ref, err = services.CheckoutBranch(conf, gitParams, checkout)
	assert.NoError(t, err, "did not expect an error")

	branch, err = testRepo.LookupBranch("master", git.BranchLocal)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, branch, "expected a branch")
	assert.Equal(t, ref, branch.Reference.Name(), "should be same reference")
}

func TestCheckoutBranchError(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.RepoBranch = "idontexist" //set branch to non existent
	testRepo := test.SetUpRepo(gitParams)
	defer testRepo.Free()
	defer test.TearDown(testRepo.Workdir())
	conf := config.NewConfig(test.TEST_CONF)
	checkout := data.NewCheckoutContext(gitParams.RepoContext)
	_, err := services.CheckoutBranch(conf, gitParams, checkout)
	assert.Error(t, err, "expected an error")

}

func TestCheckoutCommit(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(gitParams)
	defer test.TearDown(testRepo.Workdir())
	ref, err := testRepo.Head()
	assert.NoError(t, err, "did not expect an error")
	gitParams.Commit = ref.Target().String()
	gitParams.RepoType = "commit"
	checkout := data.NewCheckoutContext(gitParams.RepoContext)
	checkedOut, err := services.CheckoutCommit(conf, gitParams, checkout)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, checkedOut, "expected a commit")
}

func TestCheckoutTag(t *testing.T) {

}