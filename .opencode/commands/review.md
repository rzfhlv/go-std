---
description: Review changed Go code against the repo's modular DDD/hexagonal conventions.
agent: go-reviewer
---

Review the code specified in $ARGUMENTS. If no argument, review all uncommitted changes (staged + unstaged) plus untracked files.

Steps:
1. Run `git status --short` and `git diff HEAD` to scope the review (unless a specific file/package was named).
2. Run `make lint` and `make vet`; fold any findings into the report.
3. Apply the checklist in your instructions to every changed file — pay special attention to module boundaries and the hexagonal import direction.
4. Report findings grouped by severity with file:line references and concrete fixes.
