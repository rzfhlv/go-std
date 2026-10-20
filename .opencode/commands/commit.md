---
description: Stage changes and create a Conventional Commit for this repo.
---

Stage and commit the current changes. Request: $ARGUMENTS

Steps:
1. Run `git status --short` and `git diff HEAD` (plus `git diff --staged`) to scope what will be committed, including untracked files. If an argument names specific files, limit the commit to those.
2. Verify the diff is committable: no secrets, tokens, or `.env` files (AGENTS.md non-negotiable #7), no build artifacts (`bin/`, `coverage.out`), no unrelated changes swept in. Stop and report if it is not.
3. Classify the change with a Conventional Commit type: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `build`, `perf`, `revert`. Use the module name as scope when the change touches exactly one module (e.g. `feat(order): ...`); omit the scope otherwise.
4. Write the subject in the imperative mood, lowercase, no trailing period, ≤72 chars; add a body only when the why is not obvious from the diff. Note breaking changes with a `BREAKING CHANGE:` footer.
5. Stage only the intended files explicitly (never `git add -A`), then commit.
6. Finish with the commit hash, full message, and a one-line summary of what went in.
