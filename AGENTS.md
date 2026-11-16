# AGENTS.md

Guidance for AI coding agents working in this repository.

## Commands

- Build: `go build ./...`
- Run all tests: `go test ./...`
- Run a single test: `go test ./<pkg>/... -run '^TestName$'`
- Vet: `go vet ./...`
- Format check: `gofmt -l .` — format with `gofmt -w .`

## Conventions

- Keep tests beside the code they cover in `_test.go` files.
- When adding tooling (Makefile, lint config, CI), update this file so the commands stay accurate.

## Maintaining this file

- Only stable project facts: architecture, commands, conventions, rules.
- Never add machine-specific state (installed tools, local paths, OS, "not set up yet").
- Keep it under 100 lines; if a section grows, move it to docs/ and link it.