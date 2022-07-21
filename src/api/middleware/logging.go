package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"

	"github.com/pralolik/templgrid/src/logging"
)

func Logging(log logging.Logger) func(next http.Handler) http.Handler {
	format := "%s %d %s Remote: %s %s"

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			status := ww.Status()

			switch {
			case strings.Contains(r.RequestURI, "/health"):
				log.Info(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String())
			default:
				if status >= 400 {
					log.Error(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String())
				} else {
					log.Info(format, r.Method, status, r.RequestURI, r.RemoteAddr, time.Since(start).String())
				}
			}
		}

		return http.HandlerFunc(fn)
	}
}
