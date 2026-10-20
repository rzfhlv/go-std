---
description: Implement a plan from docs/plans/ using TDD (red -> green -> refactor), verifying every step and finishing with the full definition of done.
agent: build
---

Implement the plan: $ARGUMENTS

Use **TDD**: write the failing test first, make it pass with the smallest change, then refactor.

Steps:
1. Read `docs/plans/<name>.md` and its requirement `docs/requirements/<name>.md`. If the plan is missing, list `docs/plans/` and stop. If status is `done`, stop and say so.
2. Set the plan status to `in-progress` before starting.
3. Work through the checklist in order. For each step, run the TDD cycle:
   a. **RED** — write the test first, in the same commit-sized step as the code it drives. Derive cases from the requirement's acceptance criteria and the plan's test plan. Run it and CONFIRM IT FAILS for the right reason (a compile error counts as red; an assertion passing before implementation does not). Report the failure output.
   b. **GREEN** — write the smallest implementation that makes the test pass, following the go-conventions skill (hexagonal dependency rule, module boundaries, doc comments, `%w` error wrapping, table-driven tests). Re-read the step's files and goal. Implement only that step — no drive-by refactors, no steps ahead.
   c. **REFACTOR** — with the test green, clean up duplication, names, and structure without changing behavior; re-run the test.
   d. Run the step's verification command (and the package's full test suite for that layer); all must pass before marking the step done in the plan file (replace the step's `- [ ]` with `- ✅ `).
   e. If reality diverges from the plan (wrong assumption, better design), FIX THE PLAN FILE FIRST — adjust the step, note the change under `## Risks / open questions` — then adapt the code and continue. Never silently diverge.
4. Tests are written by you during each step, not after — never write implementation code that no test exercises. Error paths get tests in the same cycle (a domain invariant is a test case, not an afterthought).
5. As each requirement acceptance criterion becomes true, mark it done in the requirement file (replace its `- [ ]` with `- ✅ `). By the end every criterion must be ✅.
6. After all steps: run the full definition of done — `make build && make test && make lint && make vet` — and fix anything red.
7. Independent checkpoints: spawn the `test-writer` agent for any coverage gaps on changed files, and the `go-reviewer` agent on the full diff; fix every error-severity finding they report.
8. Mark the plan status `done`, and finish with: plan name, files touched, test/coverage summary, and a suggested `/commit` type (e.g. `feat(order): ...`).

