package services

import (
	"github.com/fheng/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/logger"
)

type Pusher func([]string, *git.PushOptions) error

type DefaultPusher struct {
	Refs []string
	Opts *git.PushOptions
}

func (dp *DefaultPusher) Push(push func([]string, *git.PushOptions) error) error {
	return push(dp.Refs, dp.Opts)
}

func PushToOrigin(repoContext *data.RepoContext, repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier, publicPrivateKey PublicPrivateKeyLocation, push Pusher) error {
	var (
		repoId = repoIdentifier.RepoId()
	)

	if repoContext.Local {
		logger.Logger.Info("cant push to local repo. Returning")
		return nil
	}

	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoId))
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		return err
	}

	pushOpts := &git.PushOptions{RemoteCallbacks: createRemoteCallbacks(publicPrivateKey)}
	defPush := &DefaultPusher{
		Refs: []string{ref.Name()},
		Opts: pushOpts,
	}
	if nil == push {
		push = remote.Push
	}
	return defPush.Push(push)

}
