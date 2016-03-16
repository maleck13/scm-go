package services

import (
	"regexp"
	"strings"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"
)

//todo this should be split out into two functions listTags and listBranches

func ListBranchesAndTags(branchType git.BranchType, repoLocation data.RepoLocation, repoIdentifier data.RepoIdentifier) (data.BranchesAndTags, error) {
	var (
		remoteBranches = make([]data.Remotes, 0)
		remoteTags     = make([]data.Remotes, 0)
		repoId         = repoIdentifier.RepoId()
	)
	repo, err := git.OpenRepository(repoLocation.GetRepoPath(repoId))
	if err != nil {
		return data.BranchesAndTags{}, err
	}
	branchIterator, err := repo.NewBranchIterator(branchType)
	if err != nil {
		return data.BranchesAndTags{}, err
	}

	branchIterator.ForEach(func(b *git.Branch, bt git.BranchType) error {
		name, err := b.Name()
		if err != nil {
			return err
		}
		reg := regexp.MustCompile("^[A-Za-z0-9_-]+\\/")
		name = reg.ReplaceAllString(name, "")
		target := b.Target()
		var hash string
		if nil == target {
			hash = ""
		} else {
			hash = target.String()
		}

		remoteBranches = append(remoteBranches, data.Remotes{Type: "branch", Hash: hash, Value: name})

		return nil
	})

	refIt, err := repo.NewReferenceIteratorGlob("refs/tags/*")

	if err != nil {
		return data.BranchesAndTags{}, err
	}

	tag, err := refIt.Next()
	for err == nil {
		name := strings.Replace(tag.Name(), "refs/tags/", "", 1)
		remoteTags = append(remoteTags, data.Remotes{Type: "tag", Hash: tag.Target().String(), Value: name})
		tag, err = refIt.Next()
	}

	return data.BranchesAndTags{Tags: remoteTags, Branches: remoteBranches}, nil
}
