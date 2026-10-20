# Plan: Warn when an order has no lines

- Requirement: docs/requirements/order-empty-marker.md
- Status: done

## Summary

Add an `EMPTY` column to the order CLI's `list` output, rendered as `yes` when
`app.OrderOutput.Lines` is empty and `no` otherwise. Derived entirely in the CLI
adapter (`internal/order/adapter/in/cli/cli.go`); no domain, app, port, or facade
changes (`OrderOutput` already exposes `Lines`, app/order.go:42).

## Steps

Executed TDD (red -> green -> refactor): each step names the failing test first.

- ✅ Step 1 — RED: added `TestListEmptyMarker` in `internal/order/adapter/in/cli/cli_test.go` asserting the last column is `yes` for an order with no lines and `no` for one with lines; confirmed it failed on the missing column (`header last column = "CURRENCY", want EMPTY`). GREEN: added the `EMPTY` header cell and an `emptyMarker(order)` helper used by `list`. REFACTOR: none needed. Files: `internal/order/adapter/in/cli/cli.go`, `internal/order/adapter/in/cli/cli_test.go`. Verify: `go test -race ./internal/order/adapter/in/cli/`
- ✅ Step 2 — RED: existing empty-list case already covers zero orders; confirmed passing with the new column. GREEN: no code change. Files: `internal/order/adapter/in/cli/cli_test.go`. Verify: `go test -race ./internal/order/adapter/in/cli/`
- ✅ Step 3 — usage text: no change. Divergence — `usage()` documents subcommands, not `list` columns, so there was nothing to document. Recorded under Risks.
- ✅ Step 4 — Full definition of done. Files: none. Verify: `make build && make test && make lint && make vet`

## Test plan

- domain: unchanged.
- app: unchanged.
- adapters: CLI `list` table-driven test — empty order -> `yes`, populated order -> `no`, plus the existing empty-list case. Output parsed with `strings.Fields` (tabwriter).

## Risks / open questions

- Step 3 turned out to be a no-op: `usage()` lists subcommands only and is not the place for column docs; the header row is self-documenting. No code impact.
