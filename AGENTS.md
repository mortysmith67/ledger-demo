# AGENTS.md

Guidance for AI coding agents working in this repository.

## Status

- **Ledger** is a **Go** project.
- Git repository initialized on the `main` branch.
- Go module path: `github.com/deck/ledger` (change it in `go.mod` and update this line if the canonical path differs).
- No source code or CI yet.

## Setup

- **The Go toolchain is not installed in this environment** — all `go` commands fail until it is installed. Install Go (≥1.25) and verify with `go version`.

## Commands

- Build: `go build ./...`
- Run all tests: `go test ./...`
- Run a single test: `go test ./<pkg>/... -run '^TestName$'`
- Vet: `go vet ./...`
- Format check: `gofmt -l .` — format with `gofmt -w .`

## Conventions

- Keep tests beside the code they cover in `_test.go` files.
- When adding tooling (Makefile, lint config, CI), update this file so the commands stay accurate.
