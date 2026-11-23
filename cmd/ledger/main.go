// Command ledger starts the ledger HTTP service.
//
// This bootstrap step exposes only `GET /healthz` (liveness, spec 001). The
// listen port comes from the PORT env variable; every other configuration
// variable arrives with later specs.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/mortysmith67/ledger-demo/internal/httpapi"
)

// defaultPort is the I4 exception to I3: PORT is optional and defaults to
// 8080 instead of failing fast.
const defaultPort = 8080

func main() {
	// slog JSON with an injected logger — no global (AGENTS.md → Errors
	// and logging).
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port, err := resolvePort(os.Getenv)
	if err != nil {
		// Fail fast: non-zero exit, error log names the offending
		// variable (I3 pattern, decision A).
		logger.Error("startup failed", "err", err)
		os.Exit(1)
	}

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: httpapi.NewServer(logger).Routes(),
	}

	logger.Info("ledger listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("ledger server failed", "err", err)
		os.Exit(1)
	}
}

// resolvePort resolves the listen port from the PORT env variable, read
// through the injected getenv. Unset PORT → defaultPort (I4). A malformed
// or out-of-range value aborts startup with an error naming PORT — never a
// silent fallback (decision A).
func resolvePort(getenv func(string) string) (int, error) {
	raw := getenv("PORT")
	if raw == "" {
		return defaultPort, nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid PORT %q: must be an integer between 1 and 65535", raw)
	}
	return port, nil
}
