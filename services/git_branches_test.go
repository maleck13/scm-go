package services_test

import (
	"testing"

	"github.com/maleck13/scm-go/config"
	"github.com/maleck13/scm-go/services"
	"github.com/maleck13/scm-go/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/libgit2/git2go.v23"
)

func TestListBranches(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(gitParams)
	defer testRepo.Free()
	defer test.TearDown(testRepo.Workdir())
	conf := config.NewConfig(test.TEST_CONF)

	remotes, err := services.ListBranchesAndTags(git.BranchLocal, conf, gitParams)
	assert.NoError(t, err, " did not expect an error ")
	assert.NotNil(t, remotes, "expected some remotes")
	assert.True(t, len(remotes.Branches) >= 1, "should have at least one branch")
	//check master is a branch
	var masterFound = false
	for _, v := range remotes.Branches {
		assert.Equal(t, v.Type, "branch", "expected only branches")
		if v.Value == "master" {
			masterFound = true
		}
	}
	assert.True(t, masterFound, "expected a master branch")

}
