# AGENTS.md

Guidance for AI coding agents working in this repository.

## Service

- Double-entry payment ledger core plus thin adapters (HTTP API, Stripe test
  mode, Kafka outbox); no payment business logic in adapters.
- The ledger DB is the source of truth for balances; Stripe data must be
  validated against our records before it affects the ledger.

## Stack

- chi, pgx v5, golang-migrate, franz-go, slog (JSON).
- Go module: `github.com/mortysmith67/ledger-demo`.

## Layout

- `cmd/ledger/main.go` — entry point and wiring only.
- `internal/ledger` — domain: accounts, entries, posting invariants.
- `internal/storage` — Postgres repositories and SQL.
- `internal/stripe`, `internal/kafka`, `internal/httpapi`, `internal/config`
  — thin adapters (I/O translation only).
- Dependency rule: `internal/ledger` MUST NOT import adapters or storage
  (verify imports on review); the domain compiles with fakes only.

## Commands

- `make build`, `make run`, `make test`, `make lint`, `make migrate-up`.
- Local deps: `docker compose up -d` (Postgres, Kafka).
- Stripe webhooks: `stripe listen --forward-to localhost:8080/v1/webhooks/stripe`.
- Unit tests: `go test -race ./...` (default, no Docker needed).
- Integration tests: `go test -race -tags=integration ./...` (real Postgres
  from compose). Before claiming done: `gofmt -w .`, `go vet ./...`.

## Money

- Amounts are `int64` minor units; never `float32`/`float64` for money.
- Accounts are single-currency; both legs of a posting must match the
  account currency. Mismatch is `ErrCurrencyMismatch`, never a silent cast.
- Every posting is balanced: sum of debits == sum of credits.
- Append-only ledger: corrections are new reversing entries; never UPDATE
  or DELETE existing entries.

## Transactions and concurrency

- One Postgres transaction per posting.
- Lock account rows with `SELECT ... FOR UPDATE` in ascending account-ID
  order (deterministic order prevents deadlocks).
- Enforce the floor in SQL:
  `UPDATE accounts SET balance = balance - $1 WHERE id = $2 AND balance >= $1`;
  `RowsAffected() == 0` means `ErrInsufficientFunds`.
- Never decide on a balance read outside the posting transaction.

## Idempotency

- Delivery is at-least-once everywhere; every consumer and handler MUST be
  idempotent.
- Write endpoints require `Idempotency-Key`. Store it as a column with a
  partial unique index, in the same transaction as the posting. On conflict
  (`ON CONFLICT DO NOTHING`), re-read the stored entry and return it; never
  create a second posting.
- Every Stripe API write call sends an idempotency key.

## Webhooks (Stripe test mode)

- Verify the Stripe signature (`whsec` from env) before parsing or storing.
- Ingest only: store the raw event under a UNIQUE index on `event.id`
  (dedupes redeliveries), then respond 2xx immediately.
- The worker loads our record, compares the amount, then transitions status
  and posts entries in ONE transaction.
- Never trust payload amounts alone; events may arrive out of order — check
  current state before transitioning.

## Outbox → Kafka

- The outbox row is inserted in the same transaction as the posting.
- Relay goroutine: batch `SELECT ... FOR UPDATE SKIP LOCKED` → publish →
  mark published.
- Kafka message key = account ID (per-account ordering).
- Never publish anything not written through the outbox (no dual-writes).

## Errors and logging

- Sentinel errors in `internal/ledger`; wrap with `%w`; HTTP maps them via
  `errors.Is`; never match error strings.
- slog JSON with the logger injected into service structs; error logs must
  include the `err` field; never log keys, tokens, or PII.

## Migrations

- golang-migrate, SQL files embedded via `embed`, applied on service start;
  `make migrate-up` for manual runs.
- One up/down pair per change, sequential versioning; never edit an applied
  migration — add a new pair.
- Schema keeps `NOT NULL`, CHECKs, and the idempotency UNIQUE indexes.

## Secrets

- Configuration from environment only; fail fast on missing variables.
- `.env` is local-only and gitignored; Stripe test mode only (`sk_test_…`,
  `whsec_…`); no secrets in code, tests, or logs.
- Tests use a fake Stripe server (httptest), never real keys.

## Git workflow

- Never commit to `main`. Before any change, create a branch: `feat/<short-name>`, `fix/<short-name>`, `chore/<short-name>`.
- Before starting new work, sync `main` (pull) and start new branches from it.
- Before pushing to a branch, check its PR is still open (`gh pr list --head <branch>`): commits pushed to an already-merged branch never reach `main` — re-apply them on a fresh branch from synced `main`.
- GitHub auto-deletes merged branches. After a merge: `git fetch --prune`, then `git branch -d <name>` locally. If `-d` fails, the branch has unmerged work — investigate instead of forcing.
- One task = one branch = one PR. Small commits, each builds and passes tests.
- Commit messages: Conventional Commits (`feat:`, `fix:`, `test:`, `refactor:`, `chore:`, `docs:`).
- Never rewrite published history: no `git push --force`, no `git reset --hard` on shared branches.
- Commit and/or push only on the human's explicit command, never automatically.
  Never push to `main`. After making requested changes, present the diff and wait.

## Open questions

Unresolved design questions. Do not implement these areas until each one is
closed in the specification; propose options in a PR or issue, decisions land
in the spec first.

1. Balance floor scope: `balance >= $1` is right for client-facing accounts,
   but technical accounts (clearing, the Stripe settlement account, the
   external-world account) legitimately go negative in double-entry. The spec
   must define which account kinds carry the floor.
2. `accounts.balance` is a cache over append-only entries. It may change only
   inside the same transaction as the entries it reflects, and the spec must
   define a reconciliation check that the stored balance equals the sum of
   that account's entries.
3. `Idempotency-Key` scope and payload binding: keys are unique per client (or
   per account), not globally, and the request body is part of the key's
   identity — same key with a different body returns `422`, never the stored
   response.
4. Kafka message key for a posting: a posting touches two accounts, so
   "key = account ID" is ambiguous. The spec must pick the event shape (e.g.,
   one event per account leg) before the relay is implemented.

## Maintaining this file

- Only stable project facts: architecture, commands, conventions, rules.
- Never add machine-specific state (installed tools, local paths, OS, "not set up yet").
- Keep it under 150 lines; if a section grows, move it to docs/ and link it.
