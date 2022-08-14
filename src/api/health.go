package api

import "net/http"

func (s *Server) health(rw http.ResponseWriter, r *http.Request) {
	if err := s.hc.Health(r.Context()); err != nil {
		s.log.Error("Health check failed: %v ", err)
		http.Error(rw, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
}
