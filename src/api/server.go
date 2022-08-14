package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/pralolik/templgrid/pkg"
	"github.com/pralolik/templgrid/src/api/lib/health"
	"github.com/pralolik/templgrid/src/api/middleware"
	"github.com/pralolik/templgrid/src/logging"
	"github.com/pralolik/templgrid/src/queue"
	"github.com/pralolik/templgrid/src/templatemanager"
)

const DefPort = ":8080"

type Server struct {
	hc             health.Checker
	log            logging.Logger
	httpRouter     chi.Router
	httpServer     *http.Server
	apiKey         string
	apiEnabled     bool
	previewEnabled bool
	emailStorage   *templatemanager.EmailStorage
	queue          queue.Interface
}

func NewServer(log logging.Logger, options ...Option) *Server {
	httpRouter := chi.NewRouter()

	api := Server{
		hc:  health.NewMultiChecker(),
		log: log,

		httpRouter: httpRouter,
		httpServer: &http.Server{
			Addr:              DefPort,
			Handler:           httpRouter,
			ReadHeaderTimeout: time.Minute,
		},
	}
	for _, option := range options {
		option(&api)
	}
	mwLogging := middleware.Logging(log)
	api.httpRouter.Use(mwLogging)

	if api.apiEnabled {
		api.httpRouter.Route("/email", func(r chi.Router) {
			r.Use(api.apiAuth)
			r.Use(api.jsonResponse)
			r.Post("/", api.newEmail)
		})
	}

	if api.previewEnabled {
		api.httpRouter.Route("/preview", func(r chi.Router) {
			r.Get("/", api.main)
			r.Get("/{locale}", api.locale)
			r.Get("/{locale}/{slug}", api.template)
		})
	}

	api.httpRouter.Get("/health", api.health)

	return &api
}

func (s *Server) Serve(ctx context.Context, queue queue.Interface) error {
	go s.handleShutdown(ctx)
	s.queue = queue
	s.log.Info("Server server started to serve on: '%s'", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("Server failed: %v ", err)
		var optE *net.OpError
		if errors.As(err, &optE) {
			return err
		}
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func (s *Server) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	s.log.Info("Shutting down the server!")

	killctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpServer.SetKeepAlivesEnabled(false)
	if err := s.httpServer.Shutdown(killctx); err != nil {
		s.log.Error("Panic! Failed to shutdown the HTTP server gracefully: %v ", err)
		panic(err)
	}
}

func (s *Server) apiAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("api_key")
		if apiKey != s.apiKey {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) jsonResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) sendSuccessfulResponse(rw http.ResponseWriter) {
	outgoingJSON, err := json.Marshal(pkg.SuccessfulResponse{
		Ok:      true,
		Message: "Message successfully queued",
	})
	if err != nil {
		s.sendInternalErrorResponse(rw, err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(outgoingJSON)
	if err != nil {
		s.sendInternalErrorResponse(rw, err)
		return
	}
}

func (s *Server) sendErrorValidationResponse(rw http.ResponseWriter, validationErr error) {
	outgoingJSON, err := json.Marshal(pkg.ErrorResponse{
		Ok:    false,
		Error: validationErr.Error(),
	})
	rw.WriteHeader(http.StatusBadRequest)
	if err != nil {
		s.sendInternalErrorResponse(rw, fmt.Errorf("error with marshal error validation response: %w ", err))
		return
	}
	_, err = rw.Write(outgoingJSON)
	if err != nil {
		s.sendInternalErrorResponse(rw, fmt.Errorf("error with sending error validation response: %w ", err))
		return
	}
	s.log.Info("Server validation error %d: %v ", validationErr)
}

func (s *Server) sendInternalErrorResponse(rw http.ResponseWriter, internalError error) {
	outgoingJSON, err := json.Marshal(pkg.ErrorResponse{
		Ok:    false,
		Error: "Internal error",
	})
	rw.WriteHeader(http.StatusInternalServerError)
	if err != nil {
		s.log.Error("Error with marshal error validation response: %v ", err)
		return
	}
	_, err = rw.Write(outgoingJSON)
	if err != nil {
		s.log.Error("Error with sending error validation response: %v ", err)
		return
	}
	s.log.Error("Server internal error: %v ", internalError)
}

func (s *Server) HasTemplateName(templateName string) error {
	return s.emailStorage.HasEmail(templateName)
}

type Option func(s *Server)

func WithAPI(enabled bool, apiKey string, addr string) Option {
	return func(s *Server) {
		s.apiEnabled = enabled
		s.apiKey = apiKey
		s.httpServer.Addr = addr
	}
}

func WithPreview(enabled bool, storage *templatemanager.EmailStorage) Option {
	return func(s *Server) {
		s.previewEnabled = enabled
		s.emailStorage = storage
	}
}
