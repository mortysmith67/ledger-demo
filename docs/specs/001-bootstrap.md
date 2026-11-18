# Spec: Bootstrap

- **Status:** Draft
- **Author:** Morty Smith
- **Created:** 2026-11-18

## 1. Context & goal

The repository currently has no code: only `go.mod` (module
`github.com/mortysmith67/ledger-demo`, Go 1.25), docs, and tooling config. The
stack and layout are already fixed in AGENTS.md (chi, pgx v5, golang-migrate,
franz-go, slog JSON), but there is nothing to build, run, or deploy.

This change creates the initial application skeleton: a runnable ledger
service exposing `GET /healthz`, a `Dockerfile`, and a `docker-compose.yml`
that brings up the ledger application together with Postgres. It solves the
cold-start problem: until a skeleton exists, no following spec (accounts,
postings, webhooks, outbox) has anything to build on — no running process, no
HTTP wiring, no database in the local environment.

It is worth doing first because every later spec assumes a service that
starts, answers a health check, and runs locally via `docker compose up -d`.

## 2. Scope

**In scope:**

- `cmd/ledger` + `internal/httpapi`: chi HTTP server exposing `GET /healthz`
  (liveness only, no DB checks — nothing to check yet).
- `Dockerfile` (multi-stage build) for the ledger service.
- `docker-compose.yml`: ledger service + Postgres, co-located; the app does
  **not** connect to Postgres yet.
- `Makefile`: `build`, `run`, `test`, `lint` — the targets AGENTS.md already
  references.
- GitHub Actions CI: linter + unit tests (`go test ./...`) on push/PR, plus a
  compose smoke job (AC7) and an image-pinning check (AC8).

**Explicitly out of scope:**

- Postgres connection, config loading, pgx pool → next spec (storage
  bootstrap).
- Migrations and `make migrate-up` → migrations spec (no schema exists yet —
  nothing to migrate).
- Kafka in compose, any messaging/outbox wiring → outbox spec.
- Stripe, webhooks, domain (accounts/postings) → later specs.
- Integration tests in CI → when integration tests exist (needs a Postgres
  service container then).

## 3. Contract

**Inputs:**

- `GET /healthz` — no request body, no required headers.
- The service listens on the port from the `PORT` env variable, default
  `8080`.
- No other endpoints; no Kafka topics (N/A — this spec publishes nothing).

**Outputs:**

- Success: `200 OK`, `Content-Type: application/json`:
  `{"status":"ok","time":"<RFC 3339 UTC, e.g. 2026-11-18T12:34:56Z>"}`
- No stored state, no published events.

**Error codes:**

- None — `/healthz` has no failure modes without a DB dependency. Non-GET
  methods are rejected by the chi router (`405`), unknown paths (`404`).
- Kafka/events: N/A for this spec.

## 4. Invariants

- **I1 — Liveness is dependency-free.** `GET /healthz` answers `200` from the
  running process alone; it never contacts Postgres or Kafka. Future
  readiness probes may grow dependencies; liveness must not.
- **I2 — Cold start is reproducible.** A fresh clone + `docker compose up -d`
  + `curl /healthz` succeeds with zero manual steps.
- **I3 — Fail fast on missing required env variables.** A required variable
  absent at startup aborts the process with a non-zero exit code and an error
  log naming the variable — never a degraded start with a silent fallback
  (AGENTS.md → Secrets).
- **I4 — `PORT` is the only env variable of this spec, and it is the recorded
  exception to I3:** optional, default `8080`.
- **I5 — Pinned images.** Every image in `Dockerfile` and
  `docker-compose.yml` (Go base image, Postgres) is pinned to an exact
  version; the `latest` tag is forbidden. Exact tags are picked at
  implementation and named in section 9.

No AGENTS.md money/concurrency/idempotency/webhook/outbox rule is touched:
this spec moves no money and stores nothing.

## 5. Edge cases & failures

- **Required env variable missing** → abort at startup, non-zero exit, error
  names the variable (I3). The pattern is established here even though this
  spec introduces no required variables yet.
- **`PORT` unset** → default `8080` (I4).
- **`PORT` malformed** (non-numeric or outside 1–65535) → startup aborts with
  an error naming `PORT` — never a silent default (decision A).
- **SIGTERM / SIGINT** (`docker stop`, Ctrl+C) → graceful shutdown: stop
  accepting connections, finish in-flight requests, exit `0` (decision B).
- **Duplicates / replays** — N/A: no writes, no consumers.
- **Concurrent requests** — N/A: `healthz` is stateless and read-only.
- **Postgres/Kafka unavailable** — not a failure mode here: no connections
  exist; `healthz` must not fail because of them (I1).
- **Event ordering** — N/A: no events.

## 6. Acceptance criteria

- **AC1.** Given the service is running, when `GET /healthz`, then `200` with
  `Content-Type: application/json` and body
  `{"status":"ok","time":"<RFC 3339 UTC>"}` (parsable, near current time).
- **AC2.** Given `PORT=9090`, when the service starts, then `healthz` is
  reachable on `9090`.
- **AC3.** Given `PORT` is unset, when the service starts, then `healthz` is
  reachable on `8080`.
- **AC4.** Given `PORT=not-a-port`, when the service starts, then the process
  exits non-zero and the error names `PORT` (decision A).
- **AC5.** Given the Postgres container is stopped, when `GET /healthz`, then
  `200` (I1 — liveness without dependencies).
- **AC6.** Given the service is running, when it receives `SIGTERM`, then it
  exits with code `0` (decision B).
- **AC7.** Given a fresh clone, when `docker compose up -d --build`, then
  `curl http://localhost:8080/healthz` returns the AC1 body (I2).
- **AC8.** Given the repo, when the CI image-pinning check runs, then any
  `latest` tag in `Dockerfile`/`docker-compose.yml` fails the build (I5).

## 7. Non-functional

- **Logging:** slog JSON from the first commit; the logger is injected, not a
  global. Startup and shutdown are logged; startup failures include the `err`
  field and name the offending variable — never keys, tokens, or PII
  (AGENTS.md → Errors and logging).
- Metrics: N/A — nothing to alert on yet.
- Security: covered by I5 (pinned images) and env-only config; no secrets
  exist in this spec. Container hardening (non-root user, read-only fs) →
  possible follow-up, deliberately not here.
- Performance: N/A — no money moves in this spec.

## 8. Decisions & open questions

**Decided:**

- Skeleton boundary is minimal: HTTP + `healthz` + Docker artifacts +
  Makefile + CI; no DB wiring, config module, migrations, Kafka, or Stripe
  (section 2).
- `GET /healthz` returns JSON `{"status","time"}`; `PORT` is env-provided,
  optional, default `8080` (section 3).
- Malformed `PORT` aborts startup (decision A); `SIGTERM`/`SIGINT` shut down
  gracefully with exit code `0` (decision B) (section 5).
- CI = lint + unit tests + compose smoke job (AC7) + image-pinning check
  (AC8).
- The spec fixes the **rule** (exact versions, `latest` forbidden); exact
  tags are chosen at implementation and named in the implementation plan
  (section 9).
- I3 (fail fast on required env) intentionally has no test in this spec — no
  required variables exist yet; the first one arrives together with its test.

**Open questions:**

- None. The AGENTS.md open questions (balance floor, balance-as-cache,
  `Idempotency-Key` scope, Kafka message key) are untouched by this spec.

## 9. Implementation plan

<!-- Filled by the agent AFTER the human reviews and approves sections 1–8.
Ordered steps, each shippable with a diff of ≤150 lines: migrations first,
then `internal/ledger` (domain compiles with fakes only, no adapter
imports), then `internal/storage`, then adapters. Each step states which
acceptance criteria it makes pass. -->

Not filled — completed as a separate step after sections 1–8 pass review.

## 10. Verification

| Criterion | What to break | Expected failure |
|-----------|---------------|------------------|
| AC1 | drop the `time` field / return plain text | `TestHealthzShape` fails |
| AC2 | ignore `PORT`, hardcode `8080` | `TestPortFromEnv` fails |
| AC3 | require `PORT` (fail when unset) | `TestPortDefault` fails |
| AC4 | silently fall back to `8080` on parse error | `TestPortFailsFast` fails |
| AC5 | add a DB ping into the `healthz` handler | `TestHealthzLivenessWithoutDB` fails |
| AC6 | remove signal handling from `main` | `TestGracefulShutdown` fails |
| AC7 | break the compose file (remove service / wrong port) | CI compose smoke job fails |
| AC8 | use `postgres:latest` | CI pinning check fails |
