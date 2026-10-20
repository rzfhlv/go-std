---
description: Scaffold a modular DDD/hexagonal feature or a whole new module in this Go app.
agent: build
---

Add to this modular-monolith app. Request: $ARGUMENTS

Module path is `github.com/rzfhlv/go-std`. Boundaries are convention-based (enforced by review): a module's only public surface is its facade `module.go` — never import another module's `domain/app/port/adapter` from outside that module or from `cmd/`.

## A feature inside the existing `user` module

Order matters (adapter -> app -> domain; app -> port):
1. `internal/user/domain/` — add entity/value object + invariants + sentinel errors (`errors.Is`-matchable).
2. `internal/user/port/` — declare any outbound interface needed. Interfaces only.
3. `internal/user/app/` — DTOs, inbound interface, use case implementing orchestration only. Wrap errors with `%w`.
4. `internal/user/adapter/out/` — implement the outbound port when needed.
5. `internal/user/adapter/in/{rest,grpc,cli}/` — thin adapters calling the `app` interface; map domain errors to transport errors.
6. `internal/user/module.go` — expose any new public entrypoint if needed.
7. Tests at domain, app, and adapter (rest via `httptest`) levels.

## A new module (new bounded context)

Create `internal/<module>/` with:
```
module.go
domain/doc.go       # package comment
app/doc.go          # package comment
port/               # outbound interfaces (only if needed)
adapter/in/{rest,grpc,cli}/
adapter/out/{memory,...}/
```
- Register the new module in the `depguard` rules in `.golangci.yml` (copy the `order-layers-private` rule) so its layers stay module-private.
- `module.go` is the module's ONLY public surface. Expose a constructor (`NewInMemory`) and transport entrypoints (`RegisterREST(mux *http.ServeMux)`, `RunCLI(ctx, out, args)`, `GRPCService()`).
- Never import another module's layers directly; if the new module needs another, go through that module's facade.
- Wire the new module in `cmd/rest`, `cmd/cli`, `cmd/grpc` by calling its facade — do not import module layers from `cmd/`.

Do not duplicate logic across transports; every transport calls the same `app` use case.

Verify: `make build`, `make test`, `make fmt`, `make vet` all green. Summarize files touched and example invocations.
