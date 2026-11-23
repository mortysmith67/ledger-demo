package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestRouter builds the router exactly as production does, with a
// discarded logger and — per the constructor signature — no database or
// Kafka dependency at all.
func newTestRouter() http.Handler {
	return NewServer(slog.New(slog.DiscardHandler)).Routes()
}

func doGet(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	h.ServeHTTP(rec, req)
	return rec
}

// TestHealthzShape closes AC1: GET /healthz answers 200 with
// Content-Type: application/json and body
// {"status":"ok","time":"<RFC 3339 UTC>"} where time parses as RFC 3339
// UTC and is within 5 s of the test's clock.
func TestHealthzShape(t *testing.T) {
	before := time.Now().UTC()

	rec := doGet(t, newTestRouter(), "/healthz")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	var body healthzResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v (body %q)", err, rec.Body.String())
	}
	if body.Status != "ok" {
		t.Fatalf("status field = %q, want %q", body.Status, "ok")
	}

	ts, err := time.Parse(time.RFC3339, body.Time)
	if err != nil {
		t.Fatalf("time field %q does not parse as RFC 3339: %v", body.Time, err)
	}
	if ts.Location() != time.UTC {
		t.Fatalf("time field %q is not UTC (location %v)", body.Time, ts.Location())
	}
	if d := ts.Sub(before); d < -5*time.Second || d > 5*time.Second {
		t.Fatalf("time field %q is %s from the test clock, want within ±5s", body.Time, d)
	}
}

// TestHealthzLivenessWithoutDB closes AC5 (I1): the handler is constructed
// with no database dependency and GET /healthz still answers 200.
func TestHealthzLivenessWithoutDB(t *testing.T) {
	// NewServer takes only a logger — there is no DB/Kafka parameter to
	// pass; the router must not require any backing service to answer.
	rec := doGet(t, newTestRouter(), "/healthz")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
