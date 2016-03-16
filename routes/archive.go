package routes

import (
	"bufio"
	"encoding/json"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/logger"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/services"
	"net/http"
	"os"
)

//Archive handles requests from various services looking for a zip archive of a git repo.
// It is passed a git ref which it passed on the services.ArchiveRepo method. Once a zip is created this method
// then streams it back to the response
func Archive(wr http.ResponseWriter, req *http.Request) {
	var (
		decoder = json.NewDecoder(req.Body)
		logger  = logger.Logger
		conf    = config.Conf
	)

	params, _, err := decodeAndValidate(decoder, conf)
	if nil != err {
		//we know this is an ErrorJSON so straight cast it
		HandleRouteError(req.URL, wr, err.(*data.ErrorJSON), logger)
		return
	}

	lockRepo(params)
	defer unlockRepo(params)
	path, err := services.ArchiveRepo(params.RepoContext, conf, params)
	//todo needs to possibly handle a list of files. Haven't seen this used yet. githubhandler.js 387:4
	if nil != err {
		HandleRouteError(req.URL, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}
	logger.Info("archive created ", path)
	zipFile, err := os.Open(path)
	if nil != err {
		HandleRouteError(req.URL, wr, data.NewErrorJSONUnexpectedError(err.Error()), logger)
		return
	}
	defer zipFile.Close()
	reader := bufio.NewReader(zipFile) //create a reader to stream the file to the response
	wr.Header().Set("content-type", "application/octet-stream")
	reader.WriteTo(wr) //read from the reader and write to the writer
}
