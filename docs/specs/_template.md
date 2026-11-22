# Spec: <Title>

<!--
Usage:
- Copy this file to `docs/specs/NNN-short-name.md` (sequential number, kebab-case).
  Do not edit `_template.md` when writing a spec; edit your copy.
- The human decides sections 1–8 (the agent may draft from the human's
  answers). Section 9 is written later by the planner (`/plan-spec`), only
  after sections 1–8 pass review and the status is `Accepted`; `/review-spec`
  (plan mode) reviews it, and the human then sets `Ready`.
  Section 10 is written with section 6 and drives "are these tests real?".
- If the spec exceeds ~2 pages, split the task.
- AGENTS.md is the single source of truth for project rules; this template
  references them, never restates them.
- Remove the HTML comments as you fill sections in.
-->

- **Status:** Draft <!-- Draft | In review | Accepted | Ready | Implemented — transitions in _lifecycle.md -->
- **Author:** <name>
- **Created:** <YYYY-MM-DD>

## 1. Context & goal

<!-- Why this change at all: current behavior, the problem it solves, and why
it is worth solving now. A reader who missed the discussion must understand
the goal from this section alone. -->

## 2. Scope

**In scope:**

- <!-- What this spec delivers. -->

**Explicitly out of scope:**

- <!-- The more important list: what this spec deliberately does NOT cover,
  and where that concern lives instead (another spec, follow-up, never). -->

## 3. Contract

<!-- API endpoints and/or events. Everything a caller can observe. -->

**Inputs:**

<!-- Endpoints and request bodies; headers (`Idempotency-Key` is required on
every write endpoint); Kafka topics, message key, and event payload shape
(one event per account leg vs per posting — pick and justify). -->

**Outputs:**

<!-- Response bodies, stored state, published events, status codes on success. -->

**Error codes:**

<!-- Sentinel errors from `internal/ledger` (e.g. `ErrCurrencyMismatch`,
`ErrInsufficientFunds`) mapped to HTTP status via `errors.Is`. Never match
error strings. Every error a caller can trigger is listed here. -->

## 4. Invariants

<!-- The most important section for a ledger. What must ALWAYS be true, under
any failure: crashes between steps, retries, redeliveries, concurrent
postings. If a code path can violate a statement below, the spec is wrong,
not the code.

Check every rule in AGENTS.md → Money, Transactions and concurrency,
Idempotency, Webhooks, Outbox → Kafka — and do not restate them here;
AGENTS.md stays the single source of truth. List only invariants specific to
this change:
- new guarantees this change introduces;
- existing invariants it touches (tightens, relaxes, or reinterprets — e.g.
  the balance-floor scope for a new account kind);
- which section of AGENTS.md each touched rule points to. -->

## 5. Edge cases & failures

<!-- Walk through each class explicitly; state expected behavior for each:
- duplicates: same webhook event id, same `Idempotency-Key`, replayed request;
- concurrent requests: two postings on one account, deadlocks (locking order
  is ascending account-ID), lock wait timeouts;
- timeouts/crashes: process dies after commit but before response; after
  publish but before marking the outbox row published;
- invalid input: currency mismatch, non-positive amounts, unknown accounts,
  malformed webhook signatures (reject before parsing);
- event ordering: Stripe events arriving out of order — check current state
  before transitioning, never trust payload amounts alone. -->

## 6. Acceptance criteria

<!-- Verifiable statements, Given/When/Then. Rule: one criterion = one test.
If a criterion cannot map to a single test, split it. Number them. -->

- **AC1.** Given <state>, when <action>, then <observable outcome>.

## 7. Non-functional

<!-- Fill ONLY what is relevant to this change; delete the rest.
- Logging: slog JSON, `err` field on error logs, never keys/tokens/PII.
- Metrics: what to count, what to alert on.
- Security: signature verification, secrets from env only, test-mode keys.
- Performance: latency/throughput budgets if this change moves money. -->

## 8. Decisions & open questions

**Decided:**

- <!-- Decision and why. Link to an ADR if one exists. -->

**Open questions** <!-- each must be closed (recorded in Decided) before
implementation starts; propose options with trade-offs -->

1. **<Question>:** Options: (a) …; (b) ….

## 9. Implementation plan

<!-- Written by the planner agent via `/plan-spec` — only while the status
is `Accepted` (sections 1–8 approved); after `/review-spec` (plan mode)
approves the plan, the human sets `Ready`. Format: a numbered list, one
entry per step, "Step N/M: <title>" — NO checkboxes. Each step states:
- which acceptance criteria it closes (by ID);
- which tests it adds (matching the section 10 names);
- the estimated diff in lines (≤150, build and tests green after the step);
- a review flag per the review-frequency rule in `_lifecycle.md`:
  "Review: required (<reason>)" or "Review: not required". -->

1. **Step 1/M: <title>** — closes AC…; adds tests: <section 10 test names>; diff ≈ N lines. Review: required (<reason>) | not required.

## 10. Verification

<!-- How you will convince yourself the tests are REAL and not tautologies:
for each key acceptance criterion, name the exact line/behavior to break so
the test fails. If breaking it does not fail a test, the criterion has no
test. Examples: drop `FOR UPDATE` from the locking query; invert the floor
condition `balance >= $1`; drop `ON CONFLICT DO NOTHING`; publish to Kafka
without the outbox row. Pair this with section 6 during review. -->

| Criterion | What to break | Expected failure |
|-----------|---------------|------------------|
| AC1 | <line/query/guard to break> | <test name fails> |
