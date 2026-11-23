// Command ledger starts the ledger HTTP service.
//
// This bootstrap step exposes only `GET /healthz` (liveness, spec 001). The
// listen port comes from the PORT env variable. SIGTERM/SIGINT trigger a
// graceful shutdown with a deadline; if the deadline expires, remaining
// connections are force-closed and the process still exits 0 (decision B).
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mortysmith67/ledger-demo/internal/httpapi"
)

const (
	// defaultPort is the I4 exception to I3: PORT is optional and defaults
	// to 8080 instead of failing fast.
	defaultPort = 8080

	// shutdownDeadline is the production graceful-shutdown deadline
	// (decision B). Tests inject a shorter one so they stay fast.
	shutdownDeadline = 10 * time.Second
)

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

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		logger.Error("listen failed", "err", err)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	logger.Info("ledger listening", "addr", ln.Addr().String())
	// Decision B: shutdown — graceful or forced — always exits 0; a
	// non-zero exit is reserved for startup and serve failures.
	if err := run(ln, logger, signals, shutdownDeadline, nil); err != nil {
		logger.Error("ledger server failed", "err", err)
		os.Exit(1)
	}
}

// run serves the ledger HTTP surface on ln until a signal arrives on
// signals, then shuts down within the given deadline. decorate wraps the
// router and is non-nil only in tests (the test-only slow handler).
func run(ln net.Listener, logger *slog.Logger, signals <-chan os.Signal, deadline time.Duration, decorate func(http.Handler) http.Handler) error {
	handler := httpapi.NewServer(logger).Routes()
	if decorate != nil {
		handler = decorate(handler)
	}

	srv := &http.Server{Handler: handler}

	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	select {
	case <-signals:
		logger.Info("shutdown signal received")
		return shutdown(srv, logger, deadline)
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve: %w", err)
	}
}

// shutdown drains in-flight requests with a deadline context — the primary
// path is always Shutdown, never Close. If the deadline expires, remaining
// connections are force-closed, a warning is logged with the err field,
// and the return stays nil so the process exits 0 (decision B).
func shutdown(srv *http.Server, logger *slog.Logger, deadline time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("shutdown deadline exceeded; force-closing remaining connections", "err", err)
			// Forced path only: kill whatever the drain could not
			// finish; the close error adds nothing to the warning
			// already logged.
			_ = srv.Close()
			return nil
		}
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("ledger stopped gracefully")
	return nil
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
