package httpapi

import (
	"net/http"
	"time"
)

// healthzResponse is the `GET /healthz` body (spec 001, section 3).
type healthzResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

// handleHealthz answers liveness: 200 OK, application/json, body
// {"status":"ok","time":"<RFC 3339 UTC>"}. It has no failure modes without
// a DB dependency and must never grow one (I1).
func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	resp := healthzResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	}

	s.writeJSON(w, http.StatusOK, resp)
}
