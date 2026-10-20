# Requirement: Export order list as CSV

- Status: done
- Date: 2026-10-20

## Problem

Operators of the ordering context can list orders on the REST API and in the CLI,
but there is no way to get the list into a spreadsheet. They currently copy rows
out of the terminal by hand, which is slow and error-prone.

## Solution

The CLI gains an `export` subcommand for orders that prints the order list as CSV
to standard output: one row per order with columns for id, customer, status,
subtotal, discount, total, and currency, plus a header row. The output must be
redirectable to a file with shell pipes (`order export > orders.csv`).

## Acceptance criteria

- ✅ Given orders exist, when the operator runs `order export`, then stdout is
      valid CSV with exactly one header row followed by one row per order.
- ✅ Given no orders exist, when the operator runs `order export`, then stdout
      is exactly the header row and nothing else.
- ✅ Given a customer or status value containing a comma or quote, when
      exported, then the value is correctly CSV-escaped.
- ✅ All existing tests keep passing; the new behavior has table-driven tests.

## Out of scope

- REST API CSV endpoint (CLI only for now).
- Filtering or sorting options.
- Streaming/incremental export for very large lists.

## Notes

- The order module already exposes `ListOrders` through its use cases; this
  requirement should not need new use cases or domain behavior.
- CSV formatting is a transport concern and must not leak into `domain` or `app`.
