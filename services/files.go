package services

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/maleck13/scm-go/data"
	"github.com/maleck13/scm-go/logger"
	"gopkg.in/libgit2/git2go.v23"
)

type DirList struct {
	FileList []string
	Error    error
}

type ArchiveResult struct {
	Data *bytes.Buffer
	Err  error
}

type FileResult struct {
	AbsolutePath string
	RelativePath string
}

type RepoArchive interface {
	GetRepoPath(string) string
	GetArchivePath() string
}

//Recursively reads the files in each dir below path using a new go routine for each dir and sending the file list back to the original fileRec.
//It ignores the .git directory
func RecurseReadDir(base, path string, filesRec chan DirList) {
	files, err := ioutil.ReadDir(path)
	if nil != err {
		filesRec <- DirList{nil, err}
		close(filesRec)
		return
	}
	fileNames := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			//make new channel to receive this dirs list will be closed when list returned send the list back up
			if ".git" != f.Name() {
				receiver := make(chan DirList)
				//set into next dir
				go RecurseReadDir(base, path+"/"+f.Name(), receiver)
				for dirRes := range receiver {
					filesRec <- dirRes //pipe back to main channel
				}
			}
		} else {
			cleanFile := strings.Replace(path+"/"+f.Name(), base+"/", "", 1)
			fileNames = append(fileNames, cleanFile)
		}
	}
	filesRec <- DirList{fileNames, nil}
	close(filesRec)
}

//Create or update the file. Location is the root of the git repo. The name of the file and path of the file comes from the fileParams
func CreateUpdateFile(location string, fileParams *data.RequestFile) (FileResult, error) {
	filePath := location + fileParams.Path + fileParams.Name
	shortPath := fileParams.Path + fileParams.Name
	logger.Logger.Debug("creating file with path " + filePath)
	if err := ioutil.WriteFile(filePath, []byte(fileParams.Contents), 0744); err != nil {
		return FileResult{}, err
	}
	return FileResult{filePath, shortPath[1:]}, nil
}

//Creates dir under the given repo
func CreateDir(location string, fileParams *data.RequestFile) (FileResult, error) {
	//todo you cannot create dir from ngui a the moment, so does not seem this is used.
	filePath := location + fileParams.Path + fileParams.Name
	shortPath := fileParams.Path + fileParams.Name
	logger.Logger.Debug("creating file with path " + filePath)
	if err := os.MkdirAll(filePath, 0744); err != nil {
		return FileResult{}, err
	}
	return FileResult{filePath, shortPath[1:]}, nil
}

//Facade around creating and updating either files or directories
func CreateUpdate(repoIdentity data.RepoIdentifier, repoLocation data.RepoLocation, fileParams *data.RequestFile) (FileResult, error) {

	repoPath := repoLocation.GetRepoPath(repoIdentity.RepoId())
	if fileParams.IsDirectory {
		return CreateDir(repoPath, fileParams)
	} else {
		return CreateUpdateFile(repoPath, fileParams)
	}
}

//Reads a file from the git index at the specified git ref. The git ref can be a tag/branch or a commit has If there is no path in the git index it will return a *data.ErrorJson 404 error
func ReadFile(repoIdentifier data.RepoIdentifier, repoContext *data.RepoContext, fileContext *data.FileContext, repoLocation data.RepoLocation) ([]byte, error) {
	filePath := fileContext.FullFilePath
	var commit *git.Commit

	if strings.Index(filePath, "/") == 0 {
		filePath = filePath[1:] //strings are slices so get from pos 1 to the end
	}
	rep, err := git.OpenRepository(repoLocation.GetRepoPath(repoIdentifier.RepoId()))
	if nil != err {
		return nil, err
	}
	odb, err := rep.Odb()
	if nil != err {
		return nil, err
	}

	ref, err := repoContext.GetRefValue()
	if nil != err {
		return nil, err
	}

	if repoContext.HasCommit() {
		commit, err = LookUpCommitByHash(rep, ref)
	} else {
		commit, err = lookupCommitByRef(rep, ref)
	}
	if nil != err {
		return nil, err
	}

	tree, err := commit.Tree()
	if nil != err {
		return nil, err
	}
	entry, err := tree.EntryByPath(filePath)
	if nil != err {
		return nil, data.NewErrorJSONNotFound("file not found " + filePath)
	}

	dbObj, err := odb.Read(entry.Id)
	if nil != err {
		return nil, err
	}
	return dbObj.Data(), nil
}

//List the files for passed git ref. Get the ref that can be a commit a tag or a branch. Look up the commit and get the tree of files
//Then walk this tree of files and return the list
func LsForRef(repoIdentifier data.RepoIdentifier, repoContext *data.RepoContext, repoLocation data.RepoLocation) ([]string, error) {
	rep, err := git.OpenRepository(repoLocation.GetRepoPath(repoIdentifier.RepoId()))
	if nil != err {
		return nil, err
	}

	ref, err := repoContext.GetRefValue()
	var commit *git.Commit
	if repoContext.HasCommit() {
		commit, err = LookUpCommitByHash(rep, ref)
	} else {
		commit, err = lookupCommitByRef(rep, ref)
	}
	if nil != err {
		return nil, err
	}
	tree, err := commit.Tree()
	if nil != err {
		return nil, err
	}

	fList := make([]string, 0)
	fileCB := func(path string, entry *git.TreeEntry) int {
		fList = append(fList, path+entry.Name)
		return 0
	}

	tree.Walk(fileCB)

	return fList, nil

}

//creates a new archive. Checkout out the repo at the passed ref. Writes the entries to the new archive file. returns the path to the zip
// in fh-scm it unzips to remove symlinks and re zips not really sure why we have to do that. Also dont think we need to re impliment as we are reading the file path
// and the content from the object data base todo add test with symlink files
func ArchiveRepo(repoContext *data.RepoContext, archiveParams RepoArchive, repoIdentity data.RepoIdentifier) (string, error) {
	var (
		id     string = repoIdentity.RepoId()
		commit *git.Commit
	)
	rep, err := git.OpenRepository(archiveParams.GetRepoPath(id))
	if err != nil {
		return "", err
	}
	defer rep.Free()
	odb, err := rep.Odb()
	if nil != err {
		return "", err
	}

	rand := time.Now().UnixNano()
	filePath := fmt.Sprintf("%s/%s_%d.zip", archiveParams.GetArchivePath(), id, rand)

	outputFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	ref, err := repoContext.GetRefValue()
	if err != nil {
		return "", err
	}
	archive := zip.NewWriter(outputFile)
	defer archive.Close()

	if repoContext.RepoType == "commit" {
		commit, err = LookUpCommitByHash(rep, ref)
	} else {
		commit, err = lookupCommitByRef(rep, ref)
	}
	if nil != err {
		return "", err
	}
	tree, err := commit.Tree()
	if nil != err {
		return "", err
	}
	//have our tree at the right ref location so lets walk the tree and zip up each entry
	// the underlying lib expects exit codes for errors :(
	//todo handle symlinks
	tree.Walk(func(path string, entry *git.TreeEntry) int {

		dbObj, err := odb.Read(entry.Id)
		defer dbObj.Free()
		if nil != err {
			logger.Logger.Error("err reading git entry " + err.Error())
			return 1
		}

		header := &zip.FileHeader{
			Name:         path + entry.Name,
			ModifiedDate: uint16(time.Now().UnixNano()),
			ModifiedTime: uint16(time.Now().UnixNano()),
		}

		if entry.Filemode == git.FilemodeTree {
			header.Name += "/"

		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			logger.Logger.Error("error writing archive " + err.Error())
			return 1
		}

		writer.Write(dbObj.Data())
		return 0
	})

	return filePath, nil

}

func DeleteFile(repoIdentity data.RepoIdentifier, repoLocation data.RepoLocation, fileParams *data.RequestFile) (FileResult, error) {
	repoPath := repoLocation.GetRepoPath(repoIdentity.RepoId())
	filePath := repoPath + fileParams.Path + fileParams.Name
	shortPath := fileParams.Path + fileParams.Name
	logger.Logger.Debug("deleting file with path " + filePath)
	if err := os.RemoveAll(filePath); err != nil {
		return FileResult{}, err
	}
	return FileResult{filePath, shortPath[1:]}, nil
}
