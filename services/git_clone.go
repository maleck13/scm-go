package services

import (
	"fmt"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/logger"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
)

//Clones the repo into the specified directory. A set of clone options are created. These clone options are what does a lot of the work. They take the path to the key pair
// and the branch that you want to checkout once the repo is cloned.
func CloneRepo(repoContext *data.RepoContext, repoLocator data.RepoLocation, repoIdentifier data.RepoIdentifier, publicPrivateKeyLoc PublicPrivateKeyLocation) (string, error) {
	var (
		repoId = repoIdentifier.RepoId()
	)

	cloneOpts, err := createCloneOptions(repoContext, publicPrivateKeyLoc)
	if nil != err {
		return "", err
	}
	repo, err := git.Clone(repoContext.RepoUrl, repoLocator.GetRepoPath(repoId), cloneOpts)
	if err != nil {
		return "", err
	}
	path := repo.Workdir()
	return path, nil
}

//clone options are passed to libgit2 to give it the data it needs to clone the repo. There are several callbacks available
func createCloneOptions(repoContext *data.RepoContext, keyLoc PublicPrivateKeyLocation) (*git.CloneOptions, error) {

	var logger = logger.Logger

	opts := &git.CloneOptions{
		Bare: repoContext.Bare,
		CheckoutOpts: &git.CheckoutOpts{
			Strategy: git.CheckoutForce,
			DirMode:  0744,
		},
		FetchOptions: createFetchOptions(keyLoc),
	}
	if repoContext.RepoType == "branch" {
		logger.Info("checking out ", repoContext.RepoBranch)
		opts.CheckoutBranch = repoContext.RepoBranch
	}
	return opts, nil
}

func createFetchOptions(keyLoc PublicPrivateKeyLocation) *git.FetchOptions {

	return &git.FetchOptions{
		RemoteCallbacks: createRemoteCallbacks(keyLoc),
	}
}

func createRemoteCallbacks(keyLoc PublicPrivateKeyLocation) git.RemoteCallbacks {
	var logger = logger.Logger
	return git.RemoteCallbacks{
		CredentialsCallback: func(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
			logger.Info("credentials callback " + username + " url " + url)

			ret, cred := git.NewCredSshKey(username, keyLoc.PublicKeyPath, keyLoc.PrivateKeyPath, "")
			logger.Infof("reponse int %d ", ret)
			return git.ErrorCode(ret), &cred
		},
		TransferProgressCallback: func(stats git.TransferProgress) git.ErrorCode {
			logger.Info(fmt.Sprintf("progress %d ", stats.IndexedObjects))
			return git.ErrorCode(0)
		},

		CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
			logger.Info(fmt.Sprintf(" cert check %t %s", valid, hostname))
			return git.ErrorCode(0)
		},
	}
}
