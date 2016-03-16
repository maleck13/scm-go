package routes

import (
	"encoding/json"

	"fmt"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/codegangsta/negroni"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/services"
)

func AllRoutesMiddleware(log *logrus.Logger) *negroni.Negroni {
	return negroni.New(NewHttpLogger(log))
}

type Validator interface {
	ValidateParams(...data.REQUIRED) error
}

func decodeAndValidate(decoder *json.Decoder, repoLoc data.RepoLocation) (*data.ScmRequestContext, Validator, error) {
	params := data.NewScmParams()
	if err := decoder.Decode(&params); err != nil {
		return nil, nil, data.NewErrorJSONBadRequest(err.Error())
	}

	//validate common needed params
	if err := params.ValidateCommonParams(); err != nil {
		return nil, nil, data.NewErrorJSONBadRequest(err.Error())
	}

	//returns a 204 for back compat
	if !services.RepoExists(repoLoc, params) {
		return nil, nil, data.NewErrorJsonNoContent("repo not here : " + params.RepoId())
	}

	//for some reason we send ref as a stringified json object but only sometimes... other times its a json object
	/*
			from fh-scm
			var ref = {type: gitReq.repoType, value: gitReq.repoBranch, hash: gitReq.repoCommitHash};
		    if(req.body.commit){
		      ref = {type:'commit', value: req.body.commit, hash:req.body.commit};
		    } else if(req.body.ref){
		      ref = req.body.ref;
		      if(typeof ref === 'string'){
		        ref = JSON.parse(ref);
		      }
		    }
	*/
	if v, ok := params.RefStr.(string); ok && v != "" {
		json.Unmarshal([]byte(v), params.Ref)
	} else if v, ok := params.RefStr.(map[string]string); ok {
		fmt.Println("pramas.RefStr ", v)

		params.Ref.Type = v["type"]
		params.Ref.Value = v["value"]
		params.Ref.Hash = v["hash"]

	}

	//*data.ScmRequestContext is also a Validator return this for additional validation
	return params, params, nil

}
