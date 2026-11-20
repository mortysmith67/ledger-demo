# Spec status lifecycle

A spec's `- **Status:**` line carries one of four values:
`Draft | In review | Accepted | Implemented` (see `_template.md`).
This file defines what each status means, who may set it, and how the
agents react. AGENTS.md remains the single source of general project rules;
this file covers only spec status transitions.

## Statuses

| Status | Meaning |
|---|---|
| `Draft` | Created by `/spec`; sections 1–8 interviewed, not yet reviewed. |
| `In review` | Submitted to `/review-spec`; review in progress. Optional stage — a confident Draft may go straight to `Accepted`. |
| `Accepted` | Spec review passed. Implementation may start on a `feat/NNN-*` branch. |
| `Implemented` | All acceptance criteria verified and the PR is ready; the task is done. Set by the human **before** the merge. |

## Transitions

```
            /spec (on creation)
                 │
                 ▼
              Draft ──(human)──► In review ──(human)──► Accepted ──(human)──► Implemented
                 ▲                                                  ▲              │
                 └────────────── amendments loop ───────────────────┘            PR merge
```

- `Draft` is set once, by `/spec` when the file is created.
- Every later transition (`In review`, `Accepted`, `Implemented`) is made
  **by the human only**, by editing the status line. The `/review-spec`
  verdict informs the `Accepted` decision, but the human flips the line.
- `Implemented` is set when the PR is ready and the task is done —
  **before merging it**, while the work is still on its branch. Merging is
  a git act, not a status change.
- `docs/progress/NNN.md` is a record: after a task is `Implemented` it is
  kept, never deleted. No agent may suggest deleting it.
- Abandoned specs keep `Draft` or are deleted; no other statuses exist.
  Never invent a fifth value.
- A previously `Implemented` spec may re-enter review only as a new spec
  (amendments get a new number), never by flipping the old one back.

## Closing a task

```
/review-code (APPROVE) → status Implemented → /handoff → merge
```

- `/review-code` returns `APPROVE` on the branch.
- The human flips the spec's status to `Implemented` (still before the
  merge).
- A final `/handoff` records the closing state of the task in
  `docs/progress/NNN.md`.
- The PR is merged; per AGENTS.md the branch is then cleaned up
  (`git fetch --prune; git branch -d <name>`).
- Any code change after the review — commit, amend, rebase with new
  content — invalidates the verdict: rerun `/review-code` before closing
  or merging.

## Agent rules

Agents read `Status` and react; **no agent ever edits a spec's Status**.

- **`/spec`** — writes `Draft` on creation; afterwards read-only.
- **spec-reviewer** (`/review-spec`) — accepts any of the four values in
  the header sanity check; its verdict is advice for the human, not a
  transition.
- **code reviewer** (`/review-code`) — aborts unless the status is
  `Accepted`.
- **`/handoff`** — records the current status verbatim in the progress
  file; never changes it.
- **`/resume`** — `Accepted`: normal resumption check. `Implemented`:
  one-line `CLOSED` report (nothing to resume; merge the PR if still
  open; keep the progress file as a record). `Draft` / `In review`: FAIL
  on the spec check with a pointer to `/review-spec`.
