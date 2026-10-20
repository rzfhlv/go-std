---
name: project-onboarding
description: Use when a developer is new to this repo or asks to be onboarded, oriented, or walked through the codebase — a guided tour of the modular-monolith structure, the build/run commands, the request flow, and the plan-driven development process, with hands-on checkpoints.
---

# Onboarding: github.com/rzfhlv/go-std

This repo is a **modular monolith**: one Go application, one bounded context per module,
each module built with **DDD + Hexagonal Architecture** and delivered over REST, gRPC
(placeholder), and CLI through its own adapters. This skill walks a new contributor through
it in the right order. Do it **hands-on**: run the commands, read the code, ask the
developer to make one small change before moving on.

Ground every claim in the actual files — read them, do not recite from memory.

## How to run this onboarding

Work through the stages below in order, adapting to what the developer already knows. At
each stage: explain briefly, show the real file, run the command, then checkpoint. Do not
dump all of it at once. Goal: by the end they can add a feature unaided.

Start by asking what they want: a **full tour**, a **specific area** (module, transport,
tests), or **getting productive fast**.

## Stage 1 — Get it running (10 min)

Prerequisite: Go 1.25+. Optional tooling via `make install-tools` (`golangci-lint`, `gofumpt`).

```sh
make build
make test        # -race across all packages
make run CMD=rest   # REST on :8080
make run CMD=cli ARGS="order place -id o1 -customer c1 -product p1 -name Widget -price 100 -currency USD -qty 10"
```

Checkpoint: the build is green and they can hit one REST endpoint (`curl localhost:8080/users`)
or run one CLI command. Note that `bin/cli` is **one-shot per process** — in-memory state does
not persist between invocations.

## Stage 2 — Mental model: the shape of the repo

Read `README.md` (Project structure + Architecture sections), then walk the tree:

```
cmd/{rest,cli,grpc}/main.go   # composition roots: wire module facades, own lifecycle
internal/<module>/            # one bounded context per module
  module.go                   # PUBLIC facade — the module's ONLY importable surface
  domain/                     # entities, value objects, invariants (stdlib only)
  app/                        # use cases, DTOs, inbound port interface
  port/                       # outbound interfaces only
  adapter/in/{rest,grpc,cli}/ # translate transport -> app DTOs
  adapter/out/{memory,...}/   # implement outbound ports
```

Two modules ship: **`user`** is the minimal reference slice; **`order`** is the fuller DDD
example (aggregate root, value objects, domain events, a domain service, a `Clock` port).
Copy `order` for anything non-trivial.

Checkpoint: ask them to name where a new HTTP endpoint's business rule would live — the
answer is `domain/` (invariant) or `app/` (orchestration), never `adapter/in/rest`.

## Stage 3 — Trace one request end to end

Pick the module that matches their interest and trace it with them:

- **REST create user**: `internal/user/adapter/in/rest/server.go` → `app.UserService.CreateUser`
  → `domain.NewUser` → `port.UserRepository.Save` → `adapter/out/memory`.
- **CLI place order**: `internal/order/adapter/in/cli/cli.go` → `app.OrderService.PlaceOrder`
  → `domain.NewOrder`/`AddLine` → persists + publishes events → `adapter/out/eventlog`.

Show them the README's "Request flow" mermaid diagrams alongside the code. Point out the two
rules made visible here: **all transports call the same `app` use cases**, and **domain errors
are mapped to transport errors at the adapter** (`errors.Is(err, domain.ErrX)` → status code /
CLI error).

Checkpoint: they can point to the exact line where a domain error becomes an HTTP status.

## Stage 4 — The rules that are enforced

- **Hexagonal dependency rule** (`domain` ← `app` ← adapters; `port` interfaces implemented by
  `adapter/out`). `domain/` imports only the standard library.
- **Module boundaries**: only `module.go` is public. Cross-module layer imports are rejected by
  `make lint` (depguard in `.golangci.yml`); the rest by review.
- **Non-negotiables** live in `AGENTS.md` — read it together (one module per context, facades
  only, doc comments, no `panic` in `internal/`, stdlib-first, table-driven tests, no secrets).

Load the `go-conventions` skill for the actionable reference they will use while coding.

Checkpoint: they can explain why `cmd/rest` must never `import internal/user/app`.

## Stage 5 — Make a small change (make it stick)

Have them add something tiny to build muscle memory, e.g. a new read-only REST route or a CLI
subcommand — and do it **test-first**. The order is fixed:

`domain → port → app → adapter/out → adapter/in/{rest,grpc,cli} → module.go facade → wire in cmd/`

For each layer: write the failing test (red), make it pass with the smallest change (green),
then tidy up (refactor). Run `make build && make test && make lint && make vet` after. Use
`/scaffold <request>` to see the guided recipe.

Checkpoint: they write a test that fails first, then a green build with that test passing.

## Stage 6 — The development process they will actually follow

Everything after onboarding is **plan-driven** (see `docs/README.md`):

```
docs/requirements/<name>.md  --/plan-->  docs/plans/<name>.md  --/implement-->  code + green DoD
```

1. Write/claim a requirement from `docs/requirements/_TEMPLATE.md` (acceptance criteria as
   observable outcomes).
2. `/plan <name>` → a committed, code-anchored plan whose steps each pair a failing test with
   the code it drives; review it.
3. `/implement <name>` → the agent builds step by step with **TDD** (red → green → refactor),
   verifies each step, edits the plan before code diverges, and finishes with
   `make build && make test && make lint && make vet` plus `test-writer` / `go-reviewer`
   checkpoints.
4. `/commit` for a Conventional Commit (`feat(order): ...`).

Worked example to show them: `docs/requirements/order-export-csv.md` →
`docs/plans/order-export-csv.md`.

Checkpoint: they can describe the loop, that tests come first, and that the plan file is edited
before the code when reality changes.

## Stage 7 — Where to look when stuck

| Question | Look at |
| -------- | ------- |
| What are the rules? | `AGENTS.md` |
| How is a module laid out? | `internal/order` (richest example) |
| How do I build/test/run? | `Makefile` targets; table in `README.md` |
| How does a request flow? | `README.md` request-flow diagrams + the adapter source |
| How is work planned? | `docs/README.md`, `docs/plans/` |
| Is my change allowed? | `.golangci.yml` (depguard), then `/review` |

## Ground rules for the agent running this

- Read files before describing them; quote real paths and line numbers.
- Prefer running a command over asserting its output.
- Keep stages short and interactive; defer depth until they ask.
- Never paste the whole `AGENTS.md` or README — point, summarize, and let them read.
- Finish by confirming, in their words, the two things that matter: **the dependency rule**
  and **the plan-driven loop**.
