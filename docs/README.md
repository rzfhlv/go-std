# Development process

Agentic, plan-driven development. Requirements live in this directory; plans derived
from them live in `docs/plans/`; implementation follows the plan step by step.

```
docs/requirements/<name>.md   # 1. the what (human-written, small and testable)
        │  /plan <name>
        ▼
docs/plans/<name>.md          # 2. the how (agent-written, anchored in real code)
        │  /implement <name>
        ▼
implementation + green DoD    # 3. the code (agent-built, verified at every step)
```

## 1. Write a requirement

Create `docs/requirements/<name>.md` from `docs/requirements/_TEMPLATE.md`. One
requirement = one bounded, shippable change. Name files with kebab-case
(`place-order-discount.md`).

## 2. Plan it

Run `/plan <name>`. The plan agent explores the codebase, anchors the design in
existing modules (copy the `order` module for anything non-trivial), and writes
`docs/plans/<name>.md` with ordered, file-by-file steps. Each step is one TDD cycle:
it names the failing test (red), the code that makes it pass (green), and a
verification command. **Review the plan before implementing** — it is the checkpoint
where design changes are cheap.

## 3. Implement it (TDD)

Run `/implement <name>`. The implementer works through the plan's checklist in order
using **test-driven development** — for every step:

1. **Red** — write the failing test first and confirm it fails for the right reason.
2. **Green** — write the smallest implementation that passes it.
3. **Refactor** — tidy up with the test green, then re-run.

No implementation code is written that a test does not exercise. The step is done only
when its verification command passes — then its `- [ ]` is replaced with `- ✅ ` in the
plan file. The run finishes with the full definition of done
(`make build && make test && make lint && make vet`) plus the `test-writer` and
`go-reviewer` agents as independent checkpoints.

## Conventions

- **Completion is shown with checkmarks.** A finished step in a plan, or a satisfied
  acceptance criterion in a requirement, has its `- [ ]` replaced with `- ✅ ` — the step
  text changes only in that marker, so the checklist still reads as a history of what was
  done.
- Plans are committed: they are the audit trail of why the code looks like it does.
  Mark status in the plan header (`draft` / `in-progress` / `done`) and prune plans
  older than the code that superseded them.
- A plan step that turns out wrong gets fixed **in the plan file first**, then in
  code — never silently diverge.
- Requirements describe behavior and acceptance criteria, never implementation
  (no file names or layer names in `docs/requirements/`).
