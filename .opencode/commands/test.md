---
description: Run the test suite, fix failures, and report coverage.
agent: test-writer
---

Run and stabilize the test suite for: $ARGUMENTS

Steps:
1. If an argument names a package, run `go test -race -cover ./internal/<layer>/...`; otherwise run `make test`.
2. If green, run `make cover` and report coverage plus gaps on recently changed files (`git diff --name-only`).
3. If tests fail: diagnose, fix the test or the code (prefer fixing the code unless the test asserts wrong behavior), and re-run until green. Call out any behavior changes.
4. Finish with pass/fail counts, coverage percentage, and files touched.
