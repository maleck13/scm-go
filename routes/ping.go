package routes

import "net/http"

func Ping(wr http.ResponseWriter, req *http.Request) {
	wr.Write([]byte("OK"))
}
