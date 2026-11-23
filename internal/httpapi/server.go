// Package httpapi exposes the ledger service's HTTP surface.
//
// The server is a thin adapter: it translates HTTP requests to domain
// operations. The only endpoint so far is liveness `GET /healthz`, which
// answers from the running process alone and never contacts Postgres or
// Kafka (spec 001, I1).
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Server holds the HTTP handlers. It deliberately has no database or Kafka
// dependency: liveness must not depend on any backing service (I1). The
// logger is injected, never a global (AGENTS.md → Errors and logging).
type Server struct {
	logger *slog.Logger
}

// NewServer returns a Server with the given injected logger. It panics on a
// nil logger: the logger is a required dependency, and a server without one
// must fail loudly at construction, never silently at request time
// (AGENTS.md → fail-fast convention).
func NewServer(logger *slog.Logger) *Server {
	if logger == nil {
		panic("httpapi: NewServer requires a non-nil *slog.Logger")
	}
	return &Server{logger: logger}
}

// writeJSON encodes v as JSON with the given HTTP status. It is the single
// place where response encoding and its error logging happen: the injected
// logger lives on the Server, so helpers that write responses hang off it.
func (s *Server) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		s.logger.Error("writeJSON: encode response", "err", err)
	}
}

// Routes returns the chi router with all endpoints registered. Non-GET
// methods on registered paths are rejected by the router (405), unknown
// paths return 404.
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", s.handleHealthz)
	return r
}
