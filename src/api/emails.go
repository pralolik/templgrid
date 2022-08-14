package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pralolik/templgrid/pkg"
)

func (s *Server) newEmail(rw http.ResponseWriter, r *http.Request) {
	var req pkg.TemplgridEmailEntity
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(rw, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		reqBody, readErr := io.ReadAll(r.Body)
		if readErr != nil {
			s.log.Error("Failed to set entry: %v Unable to read request body: %v ", err, readErr)
		}

		s.log.Error("Failed to set entry: failed to decode request body '%s': %v ", string(reqBody), err)
		s.sendErrorValidationResponse(rw, err)
		return
	}

	if err := req.Validate(); err != nil {
		s.log.Error("Invalid request '%v': %v ", req, err.Error())
		s.sendErrorValidationResponse(rw, err)
		return
	}

	if err := s.HasTemplateName(req.TemplateName); err != nil {
		s.log.Error("Invalid request '%v': %v ", req, err.Error())
		s.sendErrorValidationResponse(rw, err)
		return
	}

	err := s.queue.Push(&req)
	rw.Header().Set("Content-Type", "application/json")
	if err != nil {
		s.sendInternalErrorResponse(rw, err)
		return
	}
	s.sendSuccessfulResponse(rw)
	s.log.Info("New email type '%s' pushed to queue", req.TemplateName)
}
