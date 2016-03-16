package routes

import (
	"encoding/json"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/gorilla/mux"
)

func HandleRouteError(route *url.URL, wr http.ResponseWriter, err *data.ErrorJSON, logger *logrus.Logger) {
	enc := json.NewEncoder(wr)
	logger.Error("route error: ", err, " in "+route.Path)
	if err.Code > 500 {
		debug.PrintStack()
	}

	wr.WriteHeader(err.Code)
	enc.Encode(err)
}

//this is a http.Handler
type HttpPanicHandler struct {
	logger  *logrus.Logger
	handler *mux.Router
}

func RouterPanicHandler(handler *mux.Router, log *logrus.Logger) http.Handler {
	return &HttpPanicHandler{logger: log, handler: handler}
}

func (hph *HttpPanicHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//handles the panic
	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			hph.log(err)
			enc := json.NewEncoder(w)
			enc.Encode(data.NewErrorJSONUnexpectedError("failed to serve route"))

		}
	}()
	hph.handler.ServeHTTP(w, req)
}

func (hph *HttpPanicHandler) log(err interface{}) {
	hph.logger.Error("panic : ", err)
	debug.PrintStack()
}
