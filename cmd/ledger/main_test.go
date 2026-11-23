package main

import (
	"os"
	"strings"
	"testing"
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
