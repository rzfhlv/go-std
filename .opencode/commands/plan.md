---
description: Turn a requirement in docs/requirements/ into an implementation plan in docs/plans/.
agent: plan
---

Plan the implementation of the requirement: $ARGUMENTS

Steps:
1. Read `docs/requirements/<name>.md`. If missing, list `docs/requirements/` (excluding `_TEMPLATE.md`) and stop. Confirm the acceptance criteria are testable; if not, ask before planning.
2. Explore the codebase to anchor the plan in reality: locate the module(s) the change touches, read their `domain/app/port/adapter` layers and facade, and identify existing code to extend or mirror. For a new bounded context, use `internal/order` as the reference module.
3. Check `.golangci.yml` depguard rules — a new module must be registered there.
4. Design following the hexagonal rules from the go-conventions skill: invariants in `domain`, thin orchestration in `app`, outbound interfaces in `port`, transports sharing the same use cases, facade as the only public surface.
5. Write `docs/plans/<name>.md` with this shape:

```markdown
# Plan: <title>

- Requirement: docs/requirements/<name>.md
- Status: draft

## Summary
<2-3 sentences: approach and why.>

## Steps
Steps are executed with TDD (red -> green -> refactor); write each step so its test is named alongside the code it drives. Mark a finished step done by replacing its `- [ ]` with `- ✅ ` (checkmark), keeping the step text.
- [ ] Step N — <goal> (RED: <failing test `file_test.go::TestName`>; GREEN: <code that makes it pass>). Files: `<exact paths>`. Verify: `<make command or test name>`.
- [ ] Step N+1 — ...

## Test plan
- domain: <what table-driven tests cover>
- app: <use case tests incl. error paths>
- adapters: <rest via httptest, cli, out adapters>

## Risks / open questions
- <anything uncertain, with the code evidence>
```

6. Order steps by dependency (domain -> port -> app -> adapter/out -> adapter/in -> facade -> cmd wiring). Each step is one TDD cycle: it must name the failing test (RED), the code that turns it green (GREEN), exact file paths, and a verification command. Keep every step small enough to red-test and verify independently. Effects/computations that are pure get their test at the layer that owns them (invariants in domain, orchestration in app, translation in adapters).
7. Do NOT write any implementation code. Finish by summarizing the plan and reminding the user to review it before `/implement`.
