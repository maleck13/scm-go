package data

import "github.com/maleck13/scm-go/Godeps/_workspace/src/gopkg.in/libgit2/git2go.v23"

//used to marshal the response from listing tags and branches
type BranchesAndTags struct {
	Tags     []Remotes `json:"tags"`
	Branches []Remotes `json:"branches"`
}

//data structure used for adding files to a repo
type AddToRepo struct {
	RepoIdentifier RepoIdentifier
	RepoLocation   RepoLocation
	FilePaths      []string
}

func BuildAddToRepo(repoIdentifier RepoIdentifier, repoLocation RepoLocation, paths []string) AddToRepo {
	return AddToRepo{
		RepoIdentifier: repoIdentifier,
		RepoLocation:   repoLocation,
		FilePaths:      paths,
	}
}

// a type to use for different stores. Currently only disk. See KeyStore
type StoreType uint64

const (
	STORE_TYPE_DISK StoreType = 0
)

type KeyStore struct {
	StoreType StoreType
	URL       string
}

//repo identifier is an interface that has a single function that returns a string giving the name or id of a repo. An example implementation would be ScmRequestContext.RepoId()
type RepoIdentifier interface {
	RepoId() string
}

//default implementation of RepoIdentifier that takes a string and returns it from RepoId()
type RepoIdentity struct {
	identity string
}

func (ri *RepoIdentity) RepoId() string {
	return ri.identity
}

func NewRepoIdentity(identity string) RepoIdentifier {
	return &RepoIdentity{identity}
}

//returns a location to store keys
type KeyStoreLocator func() KeyStore

// interface for things that know where repos are. Implementation is *Config that has GetRepoPath(string)
type RepoLocation interface {
	GetRepoPath(string) string
}

//interface for dealing with the various kinds of checkout we have implemented by *RepoContext
type Checkout interface {
	IsBranch() bool
	IsTag() bool
	IsCommit() bool
	SetBranchType(git.BranchType)
	GetBranchType() git.BranchType
	IsLocalBranch() bool
	GetRefValue() (string, error)
}
