# AGENTS.md — agent instructions for this repository

## What this repo is

A single Go application organized as a **modular monolith**: each bounded context is a
self-contained module using **DDD + Hexagonal Architecture**, and every module is
delivered over the same three transports — REST, gRPC, CLI — through its own adapters.

```
cmd/<transport>/main.go          # composition roots: compose module facades, own lifecycle
internal/
  <module>/                      # one bounded context per module, e.g. user
    module.go                    # PUBLIC facade: constructors + transport entrypoints
    domain/                      # entities, value objects, domain errors (stdlib only)
    app/                         # use cases, DTOs, inbound port interfaces
    port/                        # outbound port interfaces only
    adapter/
      in/{rest,grpc,cli}/        # inbound adapters -> call the module's app use cases
      out/{memory,postgres,...}  # outbound adapters -> implement the module's port
```

## Module boundaries (convention — enforced by review, not the compiler)

- The **only** public surface of a module is its facade package `internal/<module>`
  (`module.go`): constructors such as `NewInMemory()` and transport entrypoints such as
  `RegisterREST`, `RunCLI`, `GRPCService`.
- Treat `internal/<module>/{domain,app,port,adapter}` as **private to the module**. Other
  modules and `cmd/` must not import them; they go through the other module's facade.
- Modules never import another module's layers. Cross-module calls use public types the
  facade exposes.
- `cmd/` imports module facades (and platform packages), never a module's layers.
- Commands never import module layers: to swap infrastructure, add a constructor to the
  facade (e.g. `user.NewPostgres(pool)`), so cmd stays decoupled.

> Go cannot enforce this without a nested `internal/` folder. Cross-module layer
> imports are rejected by `make lint` (depguard in `.golangci.yml`); the remaining
> boundaries are a documented convention checked in code review and by the
> `go-reviewer` agent. When adding a new module, register it in the depguard rules.
> If hard compiler enforcement is ever wanted, move each module's layers under
> `internal/<module>/internal/`.

## Hexagonal dependency rule (inside every module)

Imports point inward only:

```
adapter -> app -> domain
adapter -> port (implements);  app -> port (consumes)
module facade -> everything in its own module (wiring)
```

- `domain/` imports only stdlib. No port, no app, no adapter, no I/O.
- `app/` imports the module's `domain` and `port`. Never an `adapter`.
- `port/` holds outbound interfaces only. Inbound interfaces live in `app/` next to their consumer.
- `adapter/in/` translates transport to app DTOs, calls use cases, maps domain errors to transport errors.
- `adapter/out/` maps persistence/transport types to domain types at the boundary; no ORM/HTTP/GRPC types leak into `domain/` or `app/`.

DDD rules:
- Anemic domain is a smell: entities own invariants and behavior.
- Use cases orchestrate domain objects through ports; use cases hold no business rules.
- All of a module's transports drive the same `app` use cases — never fork logic per transport.

## Development process (plan-driven)

Requirements live in `docs/requirements/`; plans in `docs/plans/`. Workflow:
`/plan <name>` (requirement -> code-anchored plan) -> review plan -> `/implement <name>`
(plan -> verified implementation, TDD: red -> green -> refactor). Plans are committed and
edited before their code diverges from them. See `docs/README.md`. New contributors: the
`project-onboarding` skill is a staged tour of this repo.

## Non-negotiables

1. One module per bounded context; keep modules thin and independent. Adding a feature belongs in a module, not in `cmd/`.
2. Only a module's facade (`internal/<module>/module.go`) is public. Never leak a module's layer types through it where avoidable.
3. Every exported symbol gets a doc comment; `domain` and `app` packages have a `doc.go` package comment.
4. No `panic` in `internal/`; `cmd/` mains return non-zero on error.
5. Stdlib first; a dependency (e.g. `google.golang.org/grpc`) needs justification in the commit message.
6. TDD: write the failing test first, then the code that passes it. Table-driven tests, stdlib `testing` only, `t.Parallel()` where safe.
7. Never commit secrets, tokens, or `.env` files.

## Commands agents must use

| Task         | Command                          |
| ------------ | -------------------------------- |
| Build        | `make build`                     |
| Test         | `make test`                      |
| Coverage     | `make cover`                     |
| Lint         | `make lint`                      |
| Format       | `make fmt`                       |
| Vet          | `make vet`                       |
| Tidy         | `make tidy`                      |
| Tools setup  | `make install-tools`             |
| Run REST     | `make run CMD=rest`              |
| Run CLI      | `make run CMD=cli ARGS="user list"` |

## Definition of done

`make build && make test && make lint && make vet` all green (run `make install-tools` first if needed). Implement test-first (red -> green -> refactor). New behavior = new tests; bug fix = regression test. Conventional Commits: `feat:`, `fix:`, `chore:`, `refactor:`, `test:`, `docs:`.

## Current state

- Two modules.
  - `user` (minimal reference slice): `domain.User` -> `port.UserRepository` ->
    `app.UserService` -> `adapter/out/memory` + `adapter/in/{rest,cli}` + facade
    `user.NewInMemory()`.
  - `order` (fuller DDD example): `domain.Order` aggregate root (with `Money`/`Quantity`
    value objects, `OrderLine` entity, domain events, `DiscountPolicy` domain service) ->
    `port.{OrderRepository,EventPublisher,Clock}` -> `app.OrderService` ->
    `adapter/out/{memory,eventlog,systemclock}` + `adapter/in/{rest,cli}` + facade
    `order.NewInMemory(out io.Writer)`.
- `cmd/cli` dispatches by module name: `cli <module> <command> ...` (`user`, `order`).
  `cmd/rest` composes both module REST adapters onto one mux.
- The gRPC transport is a documented placeholder in every module
  (`internal/<module>/adapter/in/grpc`, `cmd/grpc`) pending protobuf + `google.golang.org/grpc`.

## Review checklist (go-reviewer)

- Module boundary: no cross-module layer imports; only facades are public; `cmd/` uses facades only
- Import direction inside a module respects the hexagonal rule; no ORM/protobuf/HTTP types in `domain/` or `app/`
- Domain invariants live in `domain/`; use cases stay thin orchestration
- All of a module's transports call the same use cases (no duplicated logic)
- Ports are narrow interfaces named for the consumer
- Tests cover happy path + error path; table-driven
- Errors wrapped with context at port and adapter boundaries
