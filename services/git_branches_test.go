package services_test

import (
	"testing"

	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/fheng/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
	"github.com/fheng/scm-go/config"
	"github.com/fheng/scm-go/services"
	"github.com/fheng/scm-go/test"
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
