---
description: Write a spec in docs/specs via a section-by-section interview (per docs/specs/_template.md and AGENTS.md)
agent: build
---

Write a spec for the task: **$ARGUMENTS**

If the argument is empty — first ask the human what the spec is for.
Never invent the subject of the spec yourself.

The spec author is the commit author, obtained from git:

Author: Run `git config user.name` to get the author; if empty, ask

## Preparation

1. Read `docs/specs/_template.md` and `AGENTS.md` in full. AGENTS.md is the
   single source of project rules: the spec references them, never duplicates
   them.
2. Determine the next number: list the files in `docs/specs/`, take the
   highest `NNN` + 1. Do not edit or count `_template.md`.
3. Target file: `docs/specs/NNN-<name>.md` (name in kebab-case, derived from
   the task name). Do NOT create the file until the write trigger.

## Interview — sections 1–8

- Go in order 1 → 8, **one question per message**; wait for the answer before
  moving on.
- Take the substance of each question from the HTML comments of the matching
  template section.
- Start with the header: Status: Draft, Created: today's date. Author is
  obtained above from the commit author — do not ask; if the value
  is empty, ask for the name in one question.

Handling answers:

- A vague answer → re-ask, offering **2–3 options with trade-offs** (pros and
  cons of each).
- Never invent an answer on the human's behalf and never substitute your own
  decision for theirs.
- "I don't know" → explain the options, but the decision stays with the human.

What may be skipped:

- **Sections 4 (Invariants), 5 (Edge cases), 6 (Acceptance criteria)** are
  mandatory: they cannot be skipped or marked N/A.
- The others (1–3, 7, 8) may be N/A — but only after the human explicitly
  confirms.
- Draft section 10 (Verification) together with section 6: one row per AC
  (what to break → which test must fail). ACs are numbered, one AC = one
  test; if a criterion does not map to a single test, split it.

Also:

- If the spec touches the open questions from AGENTS.md (balance floor,
  balance as a cache, `Idempotency-Key` scope, Kafka message key) — resolving
  them is part of the interview; the decision is recorded in section 8
  (Decided).
- In section 8: settled items go to Decided; unsettled ones become open
  questions with options. Open questions must be closed before implementation
  starts.
- If it is already clear the spec will exceed ~2 pages, propose splitting the
  task into several specs right away — do not wait for the interview to end.

## Finishing

1. Show the complete draft of the future file (no HTML comments; section 9
   stays as the template placeholder) and wait for the human's reaction.
2. Create or modify only `docs/specs/NNN-<name>.md` file **only if the human explicitly
   wrote the trigger — "write it"**. Discussion or edits of the
   draft are not consent. Touch no other file.
3. Do not fill in section 9 (Implementation plan). It is
   filled in a separate step after the human reviews and approves sections
   1–8.
