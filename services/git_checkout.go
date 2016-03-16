package services

import (
	"errors"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
)

type CheckoutType bool

var (
	CHECKOUT_TYPE_DETACHED CheckoutType = false
	CHECKOUT_TYPE_REGULAR  CheckoutType = true
)

//Facade around the various kinds of checkout. CheckoutBranch CheckoutTag CheckoutCommit
func Checkout(repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier, checkoutContext data.Checkout) (string, error) {
	if checkoutContext.IsBranch() {
		return CheckoutBranch(repoLocation, repoIdentifier, checkoutContext)
	} else if checkoutContext.IsTag() {
		return CheckoutTag(repoLocation, repoIdentifier, checkoutContext)
	} else if checkoutContext.IsCommit() {
		return CheckoutCommit(repoLocation, repoIdentifier, checkoutContext)
	}
	return "", errors.New("no such checkout type")
}

func checkoutHeadToRef(repo *git.Repository, ref string, detached CheckoutType) (string, error) {
	resolvedRef, err := repo.References.Lookup(ref)
	if nil != err {
		return "", err
	}
	if detached {
		if err := repo.SetHeadDetached(resolvedRef.Target()); err != nil {
			return "", err
		}
	} else if err := repo.SetHead(ref); err != nil {
		return "", err
	}

	checkoutOpts := &git.CheckoutOpts{Strategy: git.CheckoutForce}
	if err = repo.CheckoutHead(checkoutOpts); err != nil {
		return "", err
	}
	return ref, nil
}

//Checks out the branch. Head is not detached
func CheckoutBranch(repoLocator data.RepoLocation, repoIdentifier data.RepoIdentifier, checkoutContext data.Checkout) (string, error) {
	//will be something like refs/origin/somebranch
	branchRef, err := checkoutContext.GetRefValue()
	if nil != err {
		return "", err
	}
	repo, err := git.OpenRepository(repoLocator.GetRepoPath(repoIdentifier.RepoId()))
	if nil != err {
		return "", err
	}

	return checkoutHeadToRef(repo, branchRef, CHECKOUT_TYPE_REGULAR)

}

//Checks out the tag. Head is detached
func CheckoutTag(repoLocator data.RepoLocation, repoIdentifier data.RepoIdentifier, checkoutContext data.Checkout) (string, error) {
	ref, err := checkoutContext.GetRefValue()
	if nil != err {
		return "", err
	}
	repo, err := git.OpenRepository(repoLocator.GetRepoPath(repoIdentifier.RepoId()))
	if nil != err {
		return "", err
	}
	return checkoutHeadToRef(repo, ref, CHECKOUT_TYPE_DETACHED)
}

//Checks out the commit. Head is detached
func CheckoutCommit(repoLocator data.RepoLocation, repoIdentifier data.RepoIdentifier, checkoutContext data.Checkout) (string, error) {
	//ref will be the commit hash
	ref, err := checkoutContext.GetRefValue()
	if nil != err {
		return "", err
	}

	repo, err := git.OpenRepository(repoLocator.GetRepoPath(repoIdentifier.RepoId()))
	if nil != err {
		return "", err
	}

	id, err := git.NewOid(ref)
	if err != nil {
		return "", err
	}
	commit, err := repo.LookupCommit(id)
	if err != nil {
		return "", err
	}

	if err = repo.SetHeadDetached(commit.Id()); err != nil {
		return "", err
	}

	checkoutOpts := &git.CheckoutOpts{Strategy: git.CheckoutForce}
	if err = repo.CheckoutHead(checkoutOpts); err != nil {
		return "", err
	}
	return ref, nil
}
