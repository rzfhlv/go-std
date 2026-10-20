---
name: go-conventions
description: Use when writing, reviewing, or scaffolding Go code in this modular-monolith repo — covers per-module DDD + hexagonal layout, module boundaries, REST/gRPC/CLI transports, the dependency rule, error handling, testing, and the build commands agents must run.
---

# Go conventions (modular monolith: DDD + Hexagonal)

Repo: `github.com/rzfhlv/go-std`. Full policy: `AGENTS.md` at repo root.

New to the repo? Load the `project-onboarding` skill first — it is a staged, hands-on tour.

## Layout: one module per bounded context

```
cmd/{rest,grpc,cli}/main.go   # composition roots: compose module facades, own lifecycle
internal/
  <module>/                   # e.g. user
    module.go                 # PUBLIC facade ONLY: constructors + transport entrypoints
    domain/                   # entities, VOs, errors (stdlib only)
    app/                      # use cases + DTOs + inbound port interface
    port/                     # outbound interfaces only
    adapter/
      in/{rest,grpc,cli}/     # inbound adapters -> call app use cases
      out/{memory,postgres}   # outbound adapters -> implement port
```

## Boundaries (convention — enforced by review, not the compiler)

- A module's only public surface is its facade `internal/<module>/module.go`: constructors (`NewInMemory`) and transport entrypoints (`RegisterREST`, `RunCLI`, `GRPCService`).
- Treat `internal/<module>/{domain,app,port,adapter}` as private. Other modules and `cmd/` must not import them — go through the other module's facade with public types it exposes.
- `cmd/` uses facades only. To swap infrastructure, add a facade constructor (e.g. `user.NewPostgres(pool)`) — never reach into a module's layers.
- Go does not enforce this without nested `internal/`, so `make lint` (depguard) rejects cross-module layer imports and the `go-reviewer` agent checks the rest.

## Dependency rule inside a module

```
adapter -> app -> domain        app -> port        facade -> own module (wiring)
```

- `domain`: stdlib only; owns invariants (no anemic models).
- `app`: use cases via `port`; holds DTOs + inbound interface; never imports `adapter`.
- `port`: outbound interfaces only (inbound ports live in `app/`).
- `adapter/in`: translate transport -> app DTOs, map domain errors to transport errors; depend only on the `app` interface.
- `adapter/out`: implement `port`; map persistence types to domain types at the boundary; ORM types never escape.
- No `panic` in `internal/`; `cmd/` returns non-zero.

## Code rules

- Doc comment on every exported symbol starting with its name; `domain`/`app` have `doc.go`.
- Wrap errors at boundaries: `fmt.Errorf("create user: %w", err)`. Preserve sentinel domain errors for `errors.Is` mapping in adapters.
- TDD: write the failing test first (red), then the smallest code that passes it (green), then refactor. Stdlib testing only, table-driven, `t.Parallel()` where safe; happy + error path always.

## Commands

`make build` · `make test` · `make cover` · `make lint` · `make fmt` · `make vet` · `make tidy` · `make install-tools`
Run: `make run CMD=rest` / `make run CMD=grpc` / `make run CMD=cli ARGS="list"`

Done = build + test + lint + vet all green.

## Adding a feature

Within an existing module: `domain` -> `port` -> `app` -> `adapter/out` -> `adapter/in/{rest,grpc,cli}` -> facade entrypoint -> wire in `cmd/`.
New bounded context: create `internal/<module>/` with the layout above, expose a facade, compose it in the `cmd/*` roots. Register a new module in the depguard rules in `.golangci.yml`.

## Plan-driven development

Requirements: `docs/requirements/<name>.md` (template: `_TEMPLATE.md`). Plans: `docs/plans/<name>.md`.
`/plan <name>` writes the plan (anchored in real code, ordered steps each pairing a failing test with the code that drives it); `/implement <name>` executes it with TDD (red -> green -> refactor per step), verifying each with the plan's own command and finishing with build+test+lint+vet plus test-writer/go-reviewer checkpoints. Plans are edited before code silently diverges from them.
