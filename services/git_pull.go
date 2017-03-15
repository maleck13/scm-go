package services

import (
	"errors"
	"fmt"

	"github.com/maleck13/scm-go/data"
	"gopkg.in/libgit2/git2go.v23"
)

//not happiest with this imp todo poss refactor
//facade around fetch and merge which is what a git pull does
func PullRepo(repoContext *data.RepoContext, repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier, publicPrivateKey PublicPrivateKeyLocation) error {
	if err := FetchFromRemote(repoContext, repoLocation, repoIdentifier, "origin", publicPrivateKey); nil != err {
		return err
	}
	return MergeRemote(repoIdentifier, repoLocation)
}

func FetchFromRemote(repoContext *data.RepoContext, repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier, remoteName string, publicPrivateKey PublicPrivateKeyLocation) error {
	var repoId = repoIdentifier.RepoId()
	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoId))
	if err != nil {
		return err
	}

	// Locate remote
	remote, err := repo.Remotes.Lookup(remoteName)
	if err != nil {
		return err
	}
	opts := createFetchOptions(publicPrivateKey)

	err = remote.Fetch([]string{}, opts, "")
	if err != nil {
		return err
	}
	return nil

}

// merges the current checked out local branch with remote ref
//difficult one to get coverage on
func MergeRemote(repoIdentifier data.RepoIdentifier, repoLocation data.RepoLocation) error {
	var repoId = repoIdentifier.RepoId()

	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoId))
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	branch := head.Branch()
	// Get the name
	name, err := branch.Name()
	if err != nil {
		return err
	}
	//todo remove hard coded ref instead use common function
	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/" + name)
	if err != nil {
		return err
	}
	//annotated commits are ones with a message author etc (ones with meta data) as oppose to something like a tag which is not annotated
	annotatedCommit, err := repo.AnnotatedCommitFromRef(remoteBranch) //get the head commit from the origin
	if err != nil {
		return err
	}

	// Do the merge analysis
	mergeHeads := make([]*git.AnnotatedCommit, 1) //set up merge
	mergeHeads[0] = annotatedCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)
	if err != nil {
		return err
	}
	if analysis&git.MergeAnalysisUpToDate != 0 {
		//nothing to do
		return nil
	} else if analysis&git.MergeAnalysisNormal != 0 {
		// Just merge changes
		if err := repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, nil); err != nil {
			return err
		}
		// Check for conflicts
		index, err := repo.Index()
		if err != nil {
			return err
		}

		if index.HasConflicts() {
			return errors.New("Conflicts encountered. Please resolve them.")
		}

		// Make the merge commit
		sig, err := repo.DefaultSignature()
		if err != nil {
			return err
		}

		// Get Write Tree
		treeId, err := index.WriteTree()
		if err != nil {
			return err
		}

		tree, err := repo.LookupTree(treeId)
		if err != nil {
			return err
		}

		localCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			return err
		}

		remoteCommit, err := repo.LookupCommit(remoteBranch.Target())
		if err != nil {
			return err
		}

		repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)

		// Clean up
		repo.StateCleanup()
	} else if analysis&git.MergeAnalysisFastForward != 0 {
		// Fast-forward changes
		// Get remote tree
		remoteTree, err := repo.LookupTree(remoteBranch.Target())
		if err != nil {
			return err
		}

		// Checkout
		if err := repo.CheckoutTree(remoteTree, nil); err != nil {
			return err
		}

		branchRef, err := repo.References.Lookup("refs/heads/" + name)
		if err != nil {
			return err
		}
		// Point branch to the object
		branchRef.SetTarget(remoteBranch.Target(), "")
		if _, err := head.SetTarget(remoteBranch.Target(), ""); err != nil {
			return err
		}

	} else {
		return fmt.Errorf("Unexpected merge analysis result %d", analysis)
	}

	return nil

}
