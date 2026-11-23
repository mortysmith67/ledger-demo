// Package httpapi exposes the ledger service's HTTP surface.
//
// The server is a thin adapter: it translates HTTP requests to domain
// operations. The only endpoint so far is liveness `GET /healthz`, which
// answers from the running process alone and never contacts Postgres or
// Kafka (spec 001, I1).
package httpapi

import (
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

// NewServer returns a Server with the given injected logger.
func NewServer(logger *slog.Logger) *Server {
	return &Server{logger: logger}
}

// Routes returns the chi router with all endpoints registered. Non-GET
// methods on registered paths are rejected by the router (405), unknown
// paths return 404.
func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/healthz", s.handleHealthz)
	return r
}
