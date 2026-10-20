# Plan: Export order list as CSV

- Requirement: docs/requirements/order-export-csv.md
- Status: done

## Summary

Add an `export` subcommand to the order module's CLI adapter that renders `app.OrderUseCases.ListOrders` output as CSV (header + one row per order) to `c.out`. CSV is a transport concern, so the entire change lives in `internal/order/adapter/in/cli` — no domain, app, port, or facade changes are needed (`ListOrders` already exists on `app.OrderUseCases`, cli.go:57).

## Steps

- ✅ Step 1 — Add a `writeCSV(w io.Writer, orders []app.OrderOutput) error` helper in `internal/order/adapter/in/cli/cli.go` that writes the header `id,customer,status,subtotal,discount,total,currency` followed by one row per order, using `encoding/csv` (stdlib) so quoting/escaping is handled correctly. Files: `internal/order/adapter/in/cli/cli.go`. Verify: `go build ./internal/order/...`
- ✅ Step 2 — Add the `export` subcommand: dispatch in `Run` (cli.go:26), an `export(ctx)` method calling `c.orders.ListOrders(ctx)` then `writeCSV(c.out, ...)`, and extend the `usage()` text with the `export` line. Reject extra args. Files: `internal/order/adapter/in/cli/cli.go`. Verify: `go build ./internal/order/...`
- ✅ Step 3 — Table-driven tests covering: multiple orders produce header + N rows; empty list produces only the header; values containing comma/quote are CSV-escaped (round-tripped through `encoding/csv`). Files: `internal/order/adapter/in/cli/cli_test.go` (new). Verify: `go test ./internal/order/adapter/in/cli/`
- ✅ Step 4 — End-to-end CLI test: build a `CLI` with a stub `app.OrderUseCases` returning two orders, run `Run(ctx, []string{"export"})`, assert stdout parses as CSV with 3 rows; extra-arg and list-error paths return errors. Files: `internal/order/adapter/in/cli/cli_test.go`. Verify: `go test -race ./internal/order/adapter/in/cli/`
- ✅ Step 5 — Full definition of done. Files: none. Verify: `make build && make test && make lint && make vet`

## Test plan

- domain: unchanged — no domain behavior involved.
- app: unchanged — `ListOrders` already covered in `internal/order/app/order_test.go`.
- adapters: new table-driven unit tests for CSV rendering (escaping, empty, multiple) + one end-to-end `Run` test with a stub use case, mirroring the rest adapter's `httptest` approach but at the CLI boundary.

## Risks / open questions

- None identified: `encoding/csv` is stdlib (respects the stdlib-first rule), the use case interface already provides everything needed, and the depguard rules allow adapter code to import stdlib and the module's own `app` package.
- Divergence during implementation: Steps 3+4 were merged into one table-driven suite that drives CSV rendering through the public `Run(ctx, ["export"])` path with a stub use case instead of testing `writeCSV` directly. This covers everything the direct tests would (header, rows, empty, escaping via round-trip) plus the dispatch and error paths, at the cost of one extra assertion layer. No behavior difference.
