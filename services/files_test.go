package services_test

import (
	"testing"

	"github.com/fheng/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/fheng/scm-go/test"

	"fmt"
	"os"
	"sort"

	"github.com/fheng/scm-go/config"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/services"
)

type TestRepoArchiveParams struct {
	RepoPath    string
	ArchivePath string
}

func (tr *TestRepoArchiveParams) GetRepoPath(repoId string) string {
	return tr.RepoPath
}

func (tr *TestRepoArchiveParams) GetArchivePath() string {
	return tr.ArchivePath
}

func CreateFileContext(filePath string) *data.FileContext {
	return &data.FileContext{
		FullFilePath: filePath,
	}
}

func TestReadFileFromRef(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.FileContext = CreateFileContext("README")
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	refit, err := repo.NewReferenceIterator()
	assert.NoError(t, err, "did not expect an err")
	conf := config.NewConfig(test.TEST_CONF)
	name, err := refit.Next()
	assert.NoError(t, err, "did not expect an err")
	for err == nil {
		gitParams.RepoType = "branch"
		gitParams.RepoBranch = name.Name()
		t.Log(name.Shorthand())
		t.Log(name.Target().String())
		cont, fileErr := services.ReadFile(gitParams, gitParams.RepoContext, gitParams.FileContext, conf)
		assert.NoError(t, fileErr, "did not expect a file error")
		assert.Equal(t, "test", string(cont), "expected content to match")
		name, err = refit.Next()

	}
}

func TestReadFileFromRefRepoError(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.FileContext = CreateFileContext("README")
	conf := config.NewConfig(test.TEST_CONF)
	_, err := services.ReadFile(gitParams, gitParams.RepoContext, gitParams.FileContext, conf)
	assert.Error(t, err, "expected error with no repo")
}

func TestReadFileFromRefRefError(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.FileContext = CreateFileContext("README")
	gitParams.RepoBranch = "ref/heads/notthere"
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	_, err := services.ReadFile(gitParams, gitParams.RepoContext, gitParams.FileContext, conf)
	assert.Error(t, err, "expected error with no repo")
}

func TestReadFileFromRefFileError(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	gitParams.FileContext = CreateFileContext("notthere.txt")
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	_, err := services.ReadFile(gitParams, gitParams.RepoContext, gitParams.FileContext, conf)
	assert.Error(t, err, "expected error with no repo")
}

func TestLsForRefRepoError(t *testing.T) {
	conf := config.NewConfig(test.TEST_CONF)
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	_, err := services.LsForRef(gitParams, gitParams.RepoContext, conf)
	assert.Error(t, err, "expected error for no repo")
}

func TestLsForRefError(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	conf := config.NewConfig(test.TEST_CONF)
	gitParams.RepoType = "branch"
	gitParams.RepoBranch = "/refs/heads/notthere"
	_, err := services.LsForRef(gitParams, gitParams.RepoContext, conf)
	assert.Error(t, err, "expected error for no repo")
}

func TestLsForRef(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	refit, err := repo.NewReferenceIterator()
	assert.NoError(t, err, "did not expect an err")
	conf := config.NewConfig(test.TEST_CONF)
	name, err := refit.Next()
	assert.NoError(t, err, "did not expect an err")
	for err == nil {
		gitParams.RepoBranch = name.Name()
		list, fileErr := services.LsForRef(gitParams, gitParams.RepoContext, conf)
		assert.NoError(t, fileErr, "did not expect a file error")

		assert.NotNil(t, list, "expected a list of files")
		sort.Strings(list)
		t.Log(list)
		var ind = -1
		for i, v := range list {
			if v == "README" {
				ind = i
				break
			}
		}
		fmt.Printf("index is %d ", ind)
		assert.True(t, ind >= 0, "expected file to be in the list")
		name, err = refit.Next()

	}
}

func TestLsForCommit(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	conf := config.NewConfig(test.TEST_CONF)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	ref, err := repo.Head()
	assert.NoError(t, err, "did not expect an error")
	commit := ref.Target().String()
	gitParams.Ref.Type = "commit"
	gitParams.Ref.Hash = commit
	gitParams.Ref.Value = commit
	files, err := services.LsForRef(gitParams, gitParams.RepoContext, conf)
	assert.NoError(t, err, "did not expect an error")
	fmt.Println(files)
	gitParams.FileContext.RequestFile = &data.RequestFile{}
	gitParams.FileContext.RequestFile.AuthorEmail = "test@test.com"
	gitParams.FileContext.RequestFile.AuthorName = "test@test.com"
	gitParams.FileContext.RequestFile.Contents = "test2"
	gitParams.FileContext.RequestFile.IsBinary = false
	gitParams.FileContext.RequestFile.Name = "test2.text"
	gitParams.FileContext.RequestFile.Path = "/"
	fr, err := services.CreateUpdateFile(repo.Workdir(), gitParams.FileContext.RequestFile)
	assert.NoError(t, err, "did not expect an error creating file")
	retCommit, err := services.CommitFileChanges(gitParams, conf, []string{fr.RelativePath}, "test")
	assert.NoError(t, err, "did not expect an error commiting file")
	assert.NotEqual(t, commit, retCommit.Hash, "commits should not be the same")
	gitParams.Ref.Type = "commit"
	gitParams.Ref.Hash = retCommit.Hash
	gitParams.Ref.Value = retCommit.Hash
	files, err = services.LsForRef(gitParams, gitParams.RepoContext, conf)
	assert.NoError(t, err, "did not expect an error")
	fmt.Println(files)

}

func TestReadDirRecurse(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())

	receive := make(chan services.DirList)

	go services.RecurseReadDir(repo.Workdir(), repo.Workdir(), receive)

	for dList := range receive {
		assert.NoError(t, dList.Error, "no error expected from list dir")
		assert.NotNil(t, dList.FileList, "expected a list of files")
	}
}

func TestReadDirRecurseError(t *testing.T) {

	receive := make(chan services.DirList)

	go services.RecurseReadDir("/tmp/idontexist", "/tmp/idontexist", receive)

	for dList := range receive {
		assert.Error(t, dList.Error, "expected an error from list dir")
	}
}

func TestArchiveRepoBadRepo(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	_, err := services.ArchiveRepo(gitParams.RepoContext, config.Conf, gitParams)
	assert.Error(t, err, "expected error")
}

func TestArchiveRepoBadArchivePath(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	repoArchiveParams := &TestRepoArchiveParams{RepoPath: repo.Path(), ArchivePath: "/tmp/notthere/fail"}
	_, err := services.ArchiveRepo(gitParams.RepoContext, repoArchiveParams, gitParams)
	assert.Error(t, err, "expected error")
}

//todo these tests prove that the archive is created and the ref was used ok, but perhaps should have something different in each ref to assert on such as a file
func TestArchiveRepoFromCommitOk(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	ref, err := repo.Head()
	assert.NoError(t, err, "did not expect an error")
	commit := ref.Target().String()
	gitParams.Commit = commit
	gitParams.RepoType = "commit"
	gitParams.RepoBranch = commit
	path, err := services.ArchiveRepo(gitParams.RepoContext, config.Conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	defer test.TearDown(path)
	assert.NotEmpty(t, path, "expected a non empty path")
	assert.NotEmpty(t, path, "expectd a path")
	fInfo, err := os.Stat(path)
	assert.NoError(t, err, "did not expect an error")
	assert.True(t, fInfo.Size() > 0, "expected a biger file")
}

func TestArchiveRepoFromBranchOk(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	gitParams.RepoBranch = "refs/heads/master"
	path, err := services.ArchiveRepo(gitParams.RepoContext, config.Conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	defer test.TearDown(path)
	assert.NotEmpty(t, path, "expected a non empty path")
	assert.NotEmpty(t, path, "expectd a path")
	fInfo, err := os.Stat(path)
	assert.NoError(t, err, "did not expect an error")
	assert.True(t, fInfo.Size() > 0, "expected a biger file")
}

func TestArchiveRepoFromTagOk(t *testing.T) {
	gitParams := test.CreateCommonGitParams(test.TEST_APP_GUID)
	repo := test.SetUpRepo(gitParams)
	defer test.TearDown(repo.Workdir())
	gitParams.RepoBranch = test.TEST_TAG
	path, err := services.ArchiveRepo(gitParams.RepoContext, config.Conf, gitParams)
	assert.NoError(t, err, "did not expect an error")
	defer test.TearDown(path)
	assert.NotEmpty(t, path, "expected a non empty path")
	assert.NotEmpty(t, path, "expectd a path")
	fInfo, err := os.Stat(path)
	assert.NoError(t, err, "did not expect an error")
	assert.True(t, fInfo.Size() > 0, "expected a biger file")
}

//todo add symlink to repo then archive it
func TestArchiveWithSymlinks(t *testing.T) {
	t.Skip("NOT YET IMPLEMENTED")
	// create repo
	// create file in tmp
	// create symlink add to repo
	// archive
	// unzip
	// check contents of symlink file
	//os.Link() os.SymLink()
}
