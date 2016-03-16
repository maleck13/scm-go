package services

import (
	"bytes"
	"encoding/json"
	"github.com/fheng/scm-go/data"
	"github.com/fheng/scm-go/logger"
	"net/http"
	"strings"
)

//Wraps up a bump call back to millicore
type Bump struct {
	Params       *data.ScmRequestContext
	CommitHash   string
	CommandError error
	BumpTemplate string
	Enabled      bool
	BumpClient   BumpClient
}

func NewBump() *Bump {
	return &Bump{
		BumpClient: http.DefaultClient,
	}
}

type BumpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func (b *Bump) BumpUrl() string {
	url := strings.Replace(b.BumpTemplate, "CLUSTER_HOST", b.Params.RequestContext.ClusterHost, 1)
	return strings.Replace(url, "APP_GUID", b.Params.AppContext.AppGuid, 1)
}

type bumpRequest struct {
	Token     string   `json:"token"`
	CacheKey  string   `json:"cacheKey"`
	FileKey   string   `json:"fileKey"`
	Commit    string   `json:"commit"`
	Logs      []string `json:"_logs"`
	GitBranch string   `json:"gitBranch"`
	Error     string   `json:"error"`
}

func NewBumpRequest(bumpParams *Bump) map[string]interface{} {
	data := make(map[string]interface{})
	data["token"] = bumpParams.Params.RequestContext.Token
	data["cacheKey"] = bumpParams.Params.RequestContext.CacheKey
	data["fileKey"] = bumpParams.Params.RequestContext.FileKey
	data["commit"] = bumpParams.CommitHash
	data["logs"] = make([]string, 0)
	data["gitBranch"] = bumpParams.Params.RepoContext.RepoBranch
	if nil != bumpParams.CommandError {
		data["error"] = bumpParams.CommandError.Error()
	}

	return data
}

//Makes the call back to millicore to say everything is complete. It uses the a BumpClient to do this. BumpClient wraps the http call so it can be handled differently in tests
func (b *Bump) BumpMillicore() error {
	var logger = logger.Logger

	reqBody := NewBumpRequest(b)
	jsonBody, err := json.Marshal(&reqBody)
	if nil != err {
		logger.Error("error marshalling json", err)
		return err
	}

	req, err := http.NewRequest("POST", b.BumpUrl(), bytes.NewReader(jsonBody))
	if nil != err {
		logger.Error("failed to setup request", err)
		return err
	}
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	req.Header = headers
	logger.Debug("millicore is enabled ", b.Enabled)
	if b.Enabled {
		logger.Debug("calling bump millicore " + b.BumpUrl())
		_, err = b.BumpClient.Do(req)
		if nil != err {
			logger.Error("failed to bump millicore ", err)
			return err
		}
	}

	return nil

}
