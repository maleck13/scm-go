package services

import (
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
	"strings"
)

type Commit struct {
	Hash string
}

func GetHeadCommitHash(repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier) (string, error) {
	var (
		repoId = repoIdentifier.RepoId()
	)
	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoId))
	defer repo.Free()
	if err != nil {
		return "", err
	}
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}
	return ref.Target().String(), nil
}

func checkCommit(repo *git.Repository, commit string) (*data.Remotes, error) {
	resolved, err := LookUpCommitByHash(repo, commit)
	if err != nil {
		return nil, err
	}
	id := resolved.Id().String()
	gitDetails := &data.Remotes{Type: "commit", Value: id, Hash: id}
	return gitDetails, nil
}

func checkRef(repo *git.Repository, fullRef, shortRef, refType string) (*data.Remotes, error) {
	fullRef = strings.TrimSpace(fullRef)
	lRef, err := repo.References.Lookup(fullRef)
	if nil != err {
		return nil, err
	}

	hash := lRef.Target().String()
	return &data.Remotes{Type: refType, Value: shortRef, Hash: hash}, nil //scm passes back the short ref unless its a commit
}

//Checks the repo has the given ref or commit
func CheckCommitOrRef(repoContext *data.RepoContext, repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier) (*data.Remotes, error) {
	var (
		repoId = repoIdentifier.RepoId()
	)

	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoId))
	if err != nil {
		return nil, decorateError("failed to open repository ", err)
	}

	refDetails, err := repoContext.RefDetails()
	if err != nil {
		return nil, err
	}

	fullRef, err := repoContext.GetRefValue()
	if err != nil {
		return nil, err
	}

	if repoContext.HasCommit() {
		return checkCommit(repo, fullRef)
	} else {
		return checkRef(repo, fullRef, refDetails.Value, refDetails.Type)
	}

}

//looks up a ref which can be a branch or tag
func lookupCommitByRef(repo *git.Repository, ref string) (*git.Commit, error) {
	reference, err := repo.References.Lookup(ref)
	if nil != err {
		return nil, err
	}
	reference, err = reference.Resolve()
	if nil != err {
		return nil, err
	}

	oid := reference.Target()
	//tags point to a commit so need to look them up and get the targetId
	if reference.IsTag() {
		tag, err := repo.LookupTag(oid)
		if nil != err {
			return nil, err
		}
		oid = tag.TargetId()
	}

	commit, err := repo.LookupCommit(oid)
	if nil != err {
		return nil, err
	}
	return commit, nil
}

func LookUpCommitByHash(repo *git.Repository, hash string) (*git.Commit, error) {
	oid, err := git.NewOid(hash)
	if err != nil {
		return nil, err
	}
	resolved, err := repo.LookupCommit(oid)
	if err != nil {
		return nil, decorateError("failed to look up commit ", err)
	}
	return resolved, nil
}

//Adds the passed files to the repo ready for commit
func Add(addFiles data.AddToRepo) error {

	repo, err := git.OpenRepository(addFiles.RepoLocation.GetRepoPath(addFiles.RepoIdentifier.RepoId()))
	if err != nil {
		return err
	}
	index, err := repo.Index()
	if err != nil {
		return err
	}
	if err = index.AddAll(addFiles.FilePaths, git.IndexAddCheckPathspec, nil); err != nil {
		return err
	}

	_, err = index.WriteTree()
	if nil != err {
		return err
	}
	return index.Write()
}

//Commits any uncommitted changes to given repository and returns the newly created commit
func CommitChanges(repoIdentifier data.RepoIdentifier, repoLocation data.RepoLocation, commitMessage string) (*Commit, error) {

	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoIdentifier.RepoId()))
	if err != nil {
		return nil, err
	}

	index, err := repo.Index()
	if err != nil {
		return nil, decorateError("failed to get repo index during commit ", err)
	}
	treeId, err := index.WriteTree()
	if nil != err {
		return nil, decorateError("failed to write the index tree during a commit ", err)
	}
	tree, err := repo.LookupTree(treeId)
	if nil != err {
		return nil, decorateError("failed to look up the repo tree during a commit", err)
	}
	currentBranch, err := repo.Head()
	if nil != err {
		return nil, decorateError("failed to get the head of the repo", err)
	}
	currentTip, err := repo.LookupCommit(currentBranch.Target())
	if nil != err {
		return nil, decorateError("failed to look up the commit for the current branch", err)
	}
	defSig, err := repo.DefaultSignature()
	if nil != err {
		return nil, err
	}
	commitId, err := repo.CreateCommit("HEAD", defSig, defSig, commitMessage, tree, currentTip)
	if nil != err {
		return nil, decorateError("failed to create a new commit in the repo ", err)
	}

	return &Commit{commitId.String()}, nil
}

//facade arround Add and Commit
func CommitFileChanges(repoIdentifier data.RepoIdentifier, repoLocation data.RepoLocation, files []string, message string) (*Commit, error) {
	addParams := data.BuildAddToRepo(repoIdentifier, repoLocation, files)
	if err := Add(addParams); err != nil {
		return nil, err
	}
	if "" == message {
		message = "commit file changes"
	}
	return CommitChanges(repoIdentifier, repoLocation, message)
}

//todo not sure if we need this
func ResetToCommit(repoidentifier data.RepoIdentifier, repoLocation data.RepoLocation, commitHash string) error {
	return nil
}
