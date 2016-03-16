package routes

import (
	"sync"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/logger"
)

//this is a shared lock on repo actions so that only one destructive action can happen at a time on a given repo
var repo_queue map[string]*sync.Mutex = make(map[string]*sync.Mutex)

func lockRepo(repoIdentifier data.RepoIdentifier) {
	repoId := repoIdentifier.RepoId()
	if val, ok := repo_queue[repoId]; ok { //value already in map so gain the lock
		logger.Logger.Info("Lock: attempting to lock repo " + repoId)
		val.Lock()
	} else {
		//nothing there yet create the mutex and lock
		logger.Logger.Info("Lock: attempting to lock repo " + repoId)
		repo_queue[repoId] = &sync.Mutex{}
		repo_queue[repoId].Lock()
	}
	logger.Logger.Info("Lock: gained lock on repo " + repoId)
}

func unlockRepo(repoIdentifier data.RepoIdentifier) {
	repoId := repoIdentifier.RepoId()
	if val, ok := repo_queue[repoId]; ok { //value already in map so unlock
		logger.Logger.Info("Lock: unlocking repo " + repoId)
		val.Unlock()
	}
	logger.Logger.Info("Lock: unlocked repo " + repoId)
}
