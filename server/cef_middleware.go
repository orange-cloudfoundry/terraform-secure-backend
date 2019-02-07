package server

import (
	"fmt"
	"github.com/ArthurHlt/logrus-cef-formatter"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

type CEFMiddleware struct {
	logger *logrus.Logger
}

func NewCEFMiddleware(w io.Writer, version string) *CEFMiddleware {
	logger := logrus.New()
	logger.SetOutput(w)
	logger.Formatter = cef.NewCEFFormatter("orange", "terraform-secure-backend", version)
	return &CEFMiddleware{logger}
}

func (h CEFMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sw := &statusWriter{ResponseWriter: w}
		next.ServeHTTP(sw, req)
		h.logger.
			WithField("request", req.URL.Path).
			WithField("requestMethod", req.Method).
			WithField("httpStatusCode", sw.status).
			WithField("src", strings.Split(req.RemoteAddr, ":")[0]).
			WithField("xForwardedFor", strings.Replace(req.Header.Get("x-forwarded-for"), " ", "", -1)).
			Info(fmt.Sprintf("%s %s", req.Method, req.URL.Path))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}
