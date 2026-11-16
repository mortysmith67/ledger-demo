# Spec: <Title>

<!--
Usage:
- Copy this file to `docs/specs/NNN-short-name.md` (sequential number, kebab-case).
- Do not edit `_template.md` when writing a spec; edit your copy.
- Remove sections that do not apply, and remove the HTML comments as you fill them in.
- Keep implementation detail in code or migrations; the spec records intent and contracts.
-->

- **Status:** Draft <!-- Draft | In review | Accepted | Implemented -->
- **Author:** <name>
- **Created:** <YYYY-MM-DD>
- **Last updated:** <YYYY-MM-DD>

## Summary

<!-- One paragraph: what is being specified and why it matters. A reader who
knows nothing about the discussion should understand the outcome from here. -->

## Background

<!-- Current behavior, the problem being solved, and any constraints
(invariants from AGENTS.md that apply, prior decisions, related specs). -->

## Goals

- <!-- Outcome the spec must deliver. -->

## Non-goals

- <!-- Explicitly out of scope, to bound the review. -->

## Requirements

<!-- Use RFC 2119 wording: MUST / SHOULD / MAY. -->

- The system MUST …
- The system SHOULD …

## Design

### Data model

<!-- New or changed tables: columns, types (`int64` minor units for money),
NOT NULL, CHECK constraints, UNIQUE/partial-unique indexes. List the
migration pairs (up/down) required; never edit an applied migration. -->

### API surface

<!-- Endpoints, headers (e.g. `Idempotency-Key` on write endpoints),
request/response bodies, and HTTP status mapping from sentinel errors via
`errors.Is`. Never match error strings. -->

### Invariants

<!-- What must always be true after every posting:
- postings balanced: sum of debits == sum of credits;
- both legs match the account currency (`ErrCurrencyMismatch` otherwise);
- append-only ledger: corrections are new reversing entries, never UPDATE/DELETE;
- which account kinds carry the balance floor (client-facing vs technical);
- stored `accounts.balance` equals the sum of that account's entries. -->

### Transactions and concurrency

<!-- Transaction boundaries (one Postgres transaction per posting), row locks
(`SELECT ... FOR UPDATE` in ascending account-ID order), SQL-level enforcement
of the floor, and what is decided inside vs outside the transaction. -->

### Idempotency and delivery

<!-- Delivery is at-least-once everywhere. Where the idempotency key lives
(column, partial unique index, same transaction as the posting), what happens
on conflict, and how redelivered events are deduplicated. -->

### Events and outbox

<!-- Outbox row inserted in the same transaction as the posting; relay batch
behavior; Kafka message key and event shape (one event per account leg vs per
posting) with rationale; ordering guarantees. -->

### Errors and logging

<!-- New sentinel errors in `internal/ledger`; slog fields (always include
`err` for error logs). Never log keys, tokens, or PII. -->

## Testing plan

<!-- Unit tests (no Docker) and integration tests (`-tags=integration`);
invariant coverage; race behavior; Stripe webhook flow via httptest fake. -->

## Open questions

<!-- Unresolved design questions. Each MUST be closed (decision recorded in
Decisions below) before implementation lands. Propose options with trade-offs. -->

1. **<Question>:** <why it is blocking>. Options: (a) …; (b) ….

## Decisions

<!-- Decisions land here first, then in code. Append-only: never delete a row. -->

| # | Date | Decision | Rationale |
|---|------|----------|-----------|
| 1 | <YYYY-MM-DD> | <decision> | <why> |
