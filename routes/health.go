package routes

import (
	"encoding/json"
	"github.com/fheng/scm-go/data"
	"log"
	"net/http"
)

type healthRes struct {
	Http string `json:"http"`
}

func Health(wr http.ResponseWriter, req *http.Request) {
	var (
		encoder *json.Encoder
		err     error
	)
	encoder = json.NewEncoder(wr)
	if err = encoder.Encode(&healthRes{Http: "ok"}); nil != err {
		wr.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		encoder.Encode(data.NewErrorJSONUnexpectedError("failed to encode health response"))
	}

}
