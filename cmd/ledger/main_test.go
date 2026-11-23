package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// envFixture returns a getenv reading from a fixed map, so tests are
// deterministic regardless of the ambient environment.
func envFixture(env map[string]string) func(string) string {
	return func(key string) string { return env[key] }
}

// TestPortFromEnv closes AC2: PORT=9090 is honored, so the service can be
// pointed at a non-default port.
func TestPortFromEnv(t *testing.T) {
	port, err := resolvePort(envFixture(map[string]string{"PORT": "9090"}))
	if err != nil {
		t.Fatalf("resolvePort(PORT=9090) returned error: %v", err)
	}
	if port != 9090 {
		t.Fatalf("port = %d, want 9090", port)
	}
}

// TestPortDefault closes AC3 (I4): unset PORT falls back to 8080 — the one
// recorded exception to fail-fast.
func TestPortDefault(t *testing.T) {
	// No PORT key at all — not merely an empty string.
	port, err := resolvePort(envFixture(map[string]string{}))
	if err != nil {
		t.Fatalf("resolvePort(unset PORT) returned error: %v", err)
	}
	if port != 8080 {
		t.Fatalf("port = %d, want default 8080", port)
	}
}

// TestPortFailsFast closes AC4 (decision A): a malformed or out-of-range
// PORT aborts startup with an error naming PORT — never a silent fallback
// to the default.
func TestPortFailsFast(t *testing.T) {
	cases := []string{
		"not-a-port", // non-numeric
		"0",          // below range
		"65536",      // above range
		"-1",         // negative
		"12.5",       // not an integer
		" 8080",      // whitespace is not numeric
	}

	for _, raw := range cases {
		port, err := resolvePort(envFixture(map[string]string{"PORT": raw}))
		if err == nil {
			t.Fatalf("resolvePort(PORT=%q) = %d, want an error", raw, port)
		}
		if !strings.Contains(err.Error(), "PORT") {
			t.Fatalf("resolvePort(PORT=%q) error %q does not name PORT", raw, err)
		}
	}
}

// TestPortFromOsEnv pins the production wiring: os.Getenv is a valid
// getenv source for resolvePort.
func TestPortFromOsEnv(t *testing.T) {
	t.Setenv("PORT", "9091")

	port, err := resolvePort(os.Getenv)
	if err != nil {
		t.Fatalf("resolvePort(os.Getenv) returned error: %v", err)
	}
	if port != 9091 {
		t.Fatalf("port = %d, want 9091", port)
	}
}

// captureHandler is a slog.Handler that records records so tests can
// assert on what the server logged.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r.Clone())
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler      { return h }

// has reports whether a record was logged at the given level with a
// message containing substr.
func (h *captureHandler) has(level slog.Level, substr string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range h.records {
		if r.Level == level && strings.Contains(r.Message, substr) {
			return true
		}
	}
	return false
}

// startRun launches run on a fresh loopback listener. It returns the
// listener address, a channel to deliver shutdown signals on, and a
// channel carrying the run result (nil means exit 0).
func startRun(t *testing.T, logger *slog.Logger, deadline time.Duration, decorate func(http.Handler) http.Handler) (net.Addr, chan<- os.Signal, <-chan error) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	signals := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- run(ln, logger, signals, deadline, decorate) }()
	return ln.Addr(), signals, done
}

// slowHandler wraps the router with a test-only /slow endpoint sleeping d
// before answering 200. Registered only in tests (spec 001 §8).
func slowHandler(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/slow" {
				time.Sleep(d)
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("slow"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// waitUp issues GET /healthz until it answers, proving the server is
// accepting connections.
func waitUp(t *testing.T, base string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := http.Get(base + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("server never became ready: last error %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestGracefulShutdown closes AC6 (decision B): on SIGTERM the process
// path returns nil (exit 0) within the shutdown deadline and stops
// accepting connections.
func TestGracefulShutdown(t *testing.T) {
	addr, signals, done := startRun(t, slog.New(slog.NewTextHandler(io.Discard, nil)), time.Second, nil)
	base := "http://" + addr.String()

	waitUp(t, base)

	signals <- syscall.SIGTERM

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned %v after SIGTERM, want nil (exit 0)", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return within the shutdown deadline")
	}

	if _, err := http.Get(base + "/healthz"); err == nil {
		t.Fatal("server still accepting connections after shutdown")
	}
}

// TestShutdownDrainsInFlight closes AC9: a request that is in flight when
// SIGTERM arrives completes with 200 because Shutdown (not Close) drains
// it, and run still returns nil.
func TestShutdownDrainsInFlight(t *testing.T) {
	addr, signals, done := startRun(t, slog.New(slog.NewTextHandler(io.Discard, nil)), 5*time.Second, slowHandler(400*time.Millisecond))
	base := "http://" + addr.String()

	waitUp(t, base)

	type result struct {
		code int
		err  error
	}
	slowDone := make(chan result, 1)
	go func() {
		resp, err := http.Get(base + "/slow")
		if err != nil {
			slowDone <- result{err: err}
			return
		}
		defer resp.Body.Close()
		slowDone <- result{code: resp.StatusCode}
	}()

	time.Sleep(150 * time.Millisecond) // let the slow request get in flight
	signals <- syscall.SIGTERM

	select {
	case r := <-slowDone:
		if r.err != nil {
			t.Fatalf("in-flight request failed during drain: %v", r.err)
		}
		if r.code != http.StatusOK {
			t.Fatalf("in-flight request completed with %d, want 200", r.code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request did not complete during graceful drain")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned %v, want nil (exit 0)", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after draining the in-flight request")
	}
}

// TestShutdownForcedAfterDeadline closes AC10: a request outliving the
// shutdown deadline is force-closed; the process still exits 0 no later
// than deadline+1s and logs a warning about the forced close.
func TestShutdownForcedAfterDeadline(t *testing.T) {
	capture := &captureHandler{}
	addr, signals, done := startRun(t, slog.New(capture), 200*time.Millisecond, slowHandler(2*time.Second))
	base := "http://" + addr.String()

	waitUp(t, base)

	slowDone := make(chan error, 1)
	go func() {
		resp, err := http.Get(base + "/slow")
		if err == nil {
			resp.Body.Close()
		}
		slowDone <- err
	}()

	time.Sleep(100 * time.Millisecond) // let the slow request get in flight
	start := time.Now()
	signals <- syscall.SIGTERM

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned %v, want nil (decision B: exit 0)", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("run did not return after deadline expiry")
	}
	if elapsed := time.Since(start); elapsed > 1200*time.Millisecond {
		t.Fatalf("shutdown took %s, want no later than deadline+1s", elapsed)
	}
	if !capture.has(slog.LevelWarn, "force-closing") {
		t.Fatal("no warning logged about forcibly closed connections")
	}

	// The outlived request must not have completed with a response.
	select {
	case err := <-slowDone:
		if err == nil {
			t.Fatal("outlived request completed, want a forced connection failure")
		}
	case <-time.After(time.Second):
		t.Fatal("outlived request neither completed nor failed")
	}
}
