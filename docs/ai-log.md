# AI log

Record of agent mistakes in this repository: what the agent did, how the
mistake was caught, and the systemic fix that prevents a repeat. The point
is not blame — each mistake must end as a durable rule, check, or policy
change, so the same failure cannot happen silently again.

## 2026-11-19 — Machine state written into AGENTS.md

- Task: `/init` (bootstrap repository guidance).
- What the agent did: wrote machine state into AGENTS.md ("Go toolchain not
  installed", a Status section).
- How I caught it: mentor review.
- Risk: AGENTS.md drifts with one machine; every later agent inherits stale
  environment facts as if they were project rules.
- Systemic fix: "Maintaining this file" rules — stable project facts only;
  environment facts live outside the repo.

## 2026-11-19 — `git push` allowed against the stated policy

- Task: agent permissions setup (`opencode.json`).
- What the agent did: `opencode.json` allowed `git push` while AGENTS.md
  said "Do not push".
- How I caught it: mentor review.
- Risk: the agent can push against policy — including to `main`.
- Systemic fix: aligned the policy — never push to `main`; a ruleset is
  recommended as the enforcement mechanism.

## 2026-11-19 — Reviewer reconstructed a denied diff

- Task: code review run.
- What the agent did: all its shell calls were denied, so it reconstructed
  the diff by reading `.git/logs` through the read tool.
- How I caught it: not recorded; root cause was found afterwards — the
  reviewer chained commands (`&&`, `2>&1`) that did not match the allow
  patterns, so every call was denied.
- Risk: the review ran on reconstructed, possibly wrong or stale input.
- Systemic fix: exactly one git command per shell call in the reviewer, and
  the AGENTS.md "Agent conduct" rule: never reconstruct a denied input.

## 2026-11-19 — Spec review nondeterminism on spec 001

- Task: spec review of `docs/specs/001-bootstrap.md`.
- What the agent did: on the unchanged spec, run 1 found the AC5 test-level
  mismatch, run 2 missed it — so the fix never happened.
- How I caught it: comparing the two runs on the same input.
- Risk: real defects pass review by luck; their fixes never land.
- Systemic fix: explicit test-level check in spec-reviewer check 4; run the
  spec review twice for important specs.

## 2026-11-19 — Spec reviewer missed an edge case with no AC

- Task: spec review of `docs/specs/001-bootstrap.md`.
- What the agent did: missed that an edge case (in-flight requests on
  shutdown) had no covering acceptance criterion.
- How I caught it: not recorded.
- Risk: promised behavior (drain, forced close) shipped with no test.
- Systemic fix: coverage check — every invariant and edge case has an AC or
  an explicit no-test reason — plus the deadline check.

## 2026-11-23 — Planner invented a rule; plan review missed it

- Task: `/plan-spec 001` (implementation plan for `docs/specs/001-bootstrap.md`).
- What the agent did: the plan invented a rule for the pinned image tags — a
  "substitute the nearest existing exact tag and record the substitution"
  fallback the spec never contained — even though the planner instructions
  forbade inventing decisions outside the spec (the spec deliberately left
  tag picking to implementation and named the rule, not the fallback).
- How I caught it: mentor review. The plan review itself reported 0 findings.
- Risk: plans smuggle in decisions that were never approved; steps can no
  longer be anchored to approved spec decisions, and review passes them in
  silence.
- Systemic fix: every plan step carries a `Decisions:` field referencing the
  spec section each rule or choice comes from; section 9 ends with a
  mandatory `Gaps:` subsection (`None` if empty; a non-empty list means no
  plan is written — the gap must be closed in the spec first); plan review
  gains a traceability check — an unreferenced rule is MAJOR "decision
  invented by the plan", a missing or non-empty `Gaps:` is MAJOR, and
  `Gaps: None` with an unreferenced rule is MAJOR too.
