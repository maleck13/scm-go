package services

import (
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/logger"
	"os"
)

func RemoveRepo(repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier) error {
	repoPath := repoLocation.GetRepoPath(repoIdentifier.RepoId())
	return os.RemoveAll(repoPath)
}

func RepoExists(repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier) bool {
	var (
		err error
	)

	if _, err = os.Stat(repoLocation.GetRepoPath(repoIdentifier.RepoId())); err != nil {
		logger.Logger.Error("RepoExists Error :", " failed to find repo "+repoLocation.GetRepoPath(repoIdentifier.RepoId()), err.Error())
		return false
	}
	return true
}
