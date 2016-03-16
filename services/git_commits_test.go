package services_test

import (
	"fmt"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/services"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/test"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestGetHeadCommitHash(t *testing.T) {
	gitParams := data.NewScmParams()
	gitParams.AppGuid = test.TEST_COMMIT_HASH_REPO
	gitParams.ClusterName = test.TEST_CLUSTER_NAME
	gitParams.RepoUrl = test.TEST_CLUSTER_NAME

	testRepo := test.SetUpRepo(gitParams)
	conf := config.NewConfig(test.TEST_CONF)
	defer testRepo.Free()
	defer test.TearDown(testRepo.Workdir())
	cont := fmt.Sprintf("test %d ", rand.Intn(100))
	//update a file and commit it
	test.UpdateReadMe(cont, testRepo.Workdir()+"/README")
	commit, _ := test.CreateCommit(testRepo, conf)
	assert.NotNil(t, commit, "expected a commit")
	commitHash, err := services.GetHeadCommitHash(conf, gitParams)
	assert.NoError(t, err, "did not expect an err")
	assert.NotEmpty(t, commitHash, "did not expect an empty commithash")
	assert.Equal(t, commit.Id().String(), commitHash, "expected the latest commit")
	t.Log("commitHash " + commitHash)
}

func TestCheckCommitOrRefOk(t *testing.T) {
	//check branch commit
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.RepoBranch = "refs/heads/master"
	testRepo := test.SetUpRepo(gitParams)
	assert.NotNil(t, testRepo)
	defer test.TearDown(testRepo.Workdir())
	conf := config.NewConfig(test.TEST_CONF)
	rem, err := services.CheckCommitOrRef(gitParams.RepoContext, conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, rem)
	assert.NotEmpty(t, rem.Hash)

	//check a tag ref
	gitParams.RepoBranch = "atag"
	gitParams.RepoType = "tag"
	rem2, err := services.CheckCommitOrRef(gitParams.RepoContext, conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, rem2)
	assert.NotEmpty(t, rem2.Hash)

	//check a commit hash
	ref, err := testRepo.Head()
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, ref)
	commit := ref.Target().String()
	gitParams.Commit = commit
	rem3, err := services.CheckCommitOrRef(gitParams.RepoContext, conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, rem3)
	assert.NotEmpty(t, rem3.Hash)

}

func TestLookUpCommitOrRefError(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.RepoType = "commit"
	testRepo := test.SetUpRepo(gitParams)
	assert.NotNil(t, testRepo)
	defer test.TearDown(testRepo.Workdir())
	gitParams.Commit = "0bb18411fb1a83fb8b625f74330cdef78101c2cf"
	_, err := services.CheckCommitOrRef(gitParams.RepoContext, conf, gitParams)
	assert.Error(t, err, "expected an error")
}

func TestLookUpCommitByHashOk(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(gitParams)
	assert.NotNil(t, testRepo)
	defer test.TearDown(testRepo.Workdir())
	ref, err := testRepo.Head()
	assert.NoError(t, err, "not expecting an error")
	hash := ref.Target().String()
	commit, err := services.LookUpCommitByHash(testRepo, hash)
	assert.NoError(t, err, "not expecting an error")
	assert.EqualValues(t, hash, commit.Id().String(), "expected the same commit")
}

func TestResetToCommit(t *testing.T) {

}

func TestAddAndCommitFiles(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(gitParams)
	assert.NotNil(t, testRepo)
	defer test.TearDown(testRepo.Workdir())
	reqFile := &data.RequestFile{Path: "/", Name: "testFile.txt", AuthorEmail: "test@test.com", AuthorName: "test@test.com", Contents: "test", IsBinary: false, IsDirectory: false}
	fr, err := services.CreateUpdate(gitParams, conf, reqFile)
	assert.NoError(t, err, "not expecting an error")
	add := data.BuildAddToRepo(gitParams, conf, []string{fr.RelativePath})
	err = services.Add(add)
	assert.NoError(t, err, " did not expect an error")
	commit, err := services.CommitChanges(gitParams, conf, "test commit")
	assert.NoError(t, err, " did not expect an error")
	ref, err := testRepo.Head()
	assert.NoError(t, err, " did not expect an error")
	assert.EqualValues(t, commit.Hash, ref.Target().String(), "expected commits to be the same")

}

func TestCommitFileChanges(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	testRepo := test.SetUpRepo(gitParams)
	assert.NotNil(t, testRepo)
	defer test.TearDown(testRepo.Workdir())
	reqFile := &data.RequestFile{Path: "/", Name: "testFile.txt", AuthorEmail: "test@test.com", AuthorName: "test@test.com", Contents: "test", IsBinary: false, IsDirectory: false}
	fr, err := services.CreateUpdate(gitParams, conf, reqFile)
	assert.NoError(t, err, " did not expect an error")
	commit, err := services.CommitFileChanges(gitParams, conf, []string{fr.RelativePath}, "add file")
	assert.NoError(t, err, " did not expect an error")
	assert.NotNil(t, commit, "expected valid commit")
}
