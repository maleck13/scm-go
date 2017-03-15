package routes

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

type HttpLogger struct {
	*logrus.Logger
}

func (l *HttpLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	l.Debugf("Started %s %s", r.Method, r.URL.Path)

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	l.Debugf("Completed %v %s %s in %v", res.Status(), http.StatusText(res.Status()), r.URL.Path, time.Since(start))
}

func NewHttpLogger(logger *logrus.Logger) *HttpLogger {
	return &HttpLogger{logger}
}
