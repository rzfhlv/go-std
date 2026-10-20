# Requirement: Warn when an order has no lines

- Status: done
- Date: 2026-10-20

## Problem

The ordering context lets an operator list orders, but an order with no lines is
indistinguishable from a populated one in the CLI list output, so operators cannot
spot empty orders at a glance.

## Solution

The CLI `list` output gains an `EMPTY` marker column: `yes` when the order has no
lines, `no` otherwise. This is purely a presentation concern in the CLI transport —
the underlying order data is unchanged.

## Acceptance criteria

- ✅ Given an order with no lines, when the CLI lists orders, then its row shows
      `yes` in the empty column.
- ✅ Given an order with lines, when the CLI lists orders, then its row shows `no`.
- ✅ All existing CLI and package tests keep passing.

## Out of scope

- REST changes.
- Preventing empty orders (cancelling/confirming already covers that).

## Notes

- `app.OrderOutput` already carries `Lines`, so no use case or domain change is
  needed; the marker is derived in the CLI adapter.

## Traceability

- REQ-EMPTY-1: `yes`/`no` marker in the order list output.
