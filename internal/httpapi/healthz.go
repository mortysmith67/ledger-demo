package httpapi

import (
	"encoding/json"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("healthz: encode response", "err", err)
	}
}
