# Plan: Complete the user module's feature surface

- Requirement: docs/requirements/complete-user-module.md
- Status: done

## Summary

Give the `user` module a full lifecycle by wiring the already-modelled
`domain.User.Rename` (dead code today, internal/user/domain/user.go:66) into a
`RenameUser` use case, adding `DeleteUser`, and adding `CountUsers` to REST (it
exists in CLI only). Email uniqueness is enforced in `app` as a use-case outcome
(new sentinel `domain.ErrEmailTaken`), decided through a new narrow port method
`FindByEmail`. Both transports drive the same `app.UserUseCases`, so rename and
delete land in REST and CLI alike. No new module, so no `.golangci.yml` change.

## Steps

Executed TDD (red -> green -> refactor); each step names the failing test first.

- ✅ Step 1 — domain: confirm rename preserves id and email. Divergence — this is a characterization test, not a RED: `Rename` already copies the value (user.go:66) so the added id/email assertions passed on first run. Kept the assertions (they pin the invariant for later steps) and recorded the divergence under Risks. Files: `internal/user/domain/user_test.go`. Verify: `go test -race ./internal/user/domain/`
- ✅ Step 2 — domain: declare `ErrEmailTaken` (RED: `internal/user/domain/user_test.go::TestErrorSentinels` asserting `ErrEmailTaken` is non-nil and not equal to `ErrInvalidUser`; failed to compile with `undefined: domain.ErrEmailTaken`. GREEN: added the sentinel to the var block in `internal/user/domain/user.go`). Files: `internal/user/domain/user.go`, `internal/user/domain/user_test.go`. Verify: `go test -race ./internal/user/domain/`
- ✅ Step 3 — port: add `Delete` and `FindByEmail`. Files: `internal/user/port/repository.go`. Verify: `go build ./...` failed as expected with `*UserRepository does not implement port.UserRepository (missing method Delete)`.
- ✅ Step 4 — adapter/out: implement `Delete` and `FindByEmail` in memory. Divergence — moved ahead of the app steps: the app tests build their service with `memory.NewUserRepository()`, so the adapter must satisfy the widened port before the app package can compile; the original order blocked RED from running. RED: `repository_test.go::TestRepositoryDelete` and `::TestRepositoryFindByEmail` failed to compile with `missing method Delete`. GREEN: implemented both over the existing map, normalizing the lookup email the same way `domain.NewUser` stores it. Files: `internal/user/adapter/out/memory/repository.go`, `internal/user/adapter/out/memory/repository_test.go`. Verify: `go test -race ./internal/user/adapter/out/memory/` and `go build ./...`
- ✅ Step 5 — app: `RenameUser` and `DeleteUser`. RED: `TestRenameUser`/`TestDeleteUser` failed to compile (`svc.RenameUser undefined`, `svc.DeleteUser undefined`). GREEN: added both to `UserUseCases` and `UserService`, reusing `domain.User.Rename`. Files: `internal/user/app/user.go`, `internal/user/app/user_test.go`. Verify: `go test -race ./internal/user/app/`
- ✅ Step 6 — app: reject duplicate emails in `CreateUser`. RED: `TestCreateUserDuplicateEmail` failed with `CreateUser() error = <nil>, want email already in use` for the exact and case/space variants. GREEN: `CreateUser` calls `repo.FindByEmail`, mapping any hit to `domain.ErrEmailTaken` and treating `ErrUserNotFound` as free. Also fixed `TestCountUsers`, which seeded duplicate emails and would now correctly fail. Files: `internal/user/app/user.go`, `internal/user/app/user_test.go`. Verify: `go test -race ./internal/user/app/`
- ✅ Step 7 — adapter/in/cli: `rename` and `delete` subcommands. RED: new `internal/user/adapter/in/cli/cli_test.go` (stub `UserUseCases`) failed with `unknown command "rename"` / `unknown command "delete"`; also added an error-path test for `get` and an end-to-end test against the memory adapter. GREEN: added `case "rename"` / `case "delete"`, their handlers, and usage lines. Files: `internal/user/adapter/in/cli/cli.go`, `internal/user/adapter/in/cli/cli_test.go`. Verify: `go test -race ./internal/user/adapter/in/cli/`
- ✅ Step 8 — adapter/in/rest: rename, delete, and count routes. RED: extended `server_test.go` — PATCH 405, DELETE 405, `/users/count` 404 (captured by `/users/{id}`), duplicate email 500 (unmapped sentinel). GREEN: registered `GET /users/count`, `PATCH /users/{id}`, `DELETE /users/{id}`; added a `renameUserRequest` DTO and mapped `ErrEmailTaken` -> 409 in `writeDomainError`. Files: `internal/user/adapter/in/rest/server.go`, `internal/user/adapter/in/rest/server_test.go`. Verify: `go test -race ./internal/user/adapter/in/rest/`
- ✅ Step 9 — facade wiring. No change required, as predicted: the facade exposes `app.UserUseCases`, so the new methods flow through automatically. Verified by `make build` and `make run CMD=cli ARGS="user count"` printing `0`. Files: none. Verify: `make build && make run CMD=cli ARGS="user count"`
- ✅ Step 10 — full definition of done + independent checkpoints. `make build && make test && make lint && make vet` all green. `test-writer` added error-path tests (app repo-error and save-error branches, REST invalid-body and upstream-error, CLI flag-parse errors); coverage now app 92.3%, rest 83.3%, memory 76.7%, domain 100%. `go-reviewer` reported no error-severity findings across two passes; its warnings were addressed: domain.NormalizeEmail extracted and shared, uniqueness limitation documented on the port and CreateUser, CLI missing-id usage guards, /users/count precedence pinned by test. Files: `internal/user/adapter/in/cli/cli.go` (guards), plus the tests above. Verify: `make build && make test && make lint && make vet`

## Test plan

- domain: `NewUser` (existing table extended for email normalization), `Rename`
  preserves id/email and rejects blank, `ErrEmailTaken` distinctness.
- app: `RenameUser` happy/blank/not-found; `DeleteUser` happy/second-delete;
  `CreateUser` duplicate-email -> `ErrEmailTaken` with nothing stored. Table-driven,
  backed by `memory.NewUserRepository()`.
- adapters: REST via `httptest` (200/204/400/404/409); CLI via a stub
  `UserUseCases` mirroring the order module's `stubOrders`
  (internal/order/adapter/in/cli/cli_test.go:15); out/memory gains its own
  `repository_test.go` (it has none today).

## Risks / open questions

- Step 1 was a characterization test rather than a RED: `Rename` already preserved
  id and email, so the new assertions passed immediately. The assertions were kept
  because they lock the invariant in before `RenameUser` reaches them.
- Step 4 was moved ahead of the app steps (originally Step 6): the app tests build
  their service with `memory.NewUserRepository()`, so the adapter must satisfy the
  widened port before the app package compiles. Plan reordered accordingly.
- Design decisions taken with the plan author's recommendation, since no explicit
  answer was given: (1) `ErrEmailTaken` lives in `domain` for consistency with the
  other sentinels, though only `app` produces it; (2) rename is `PATCH /users/{id}`;
  (3) uniqueness is an `app`-level check via a new `FindByEmail` port method.
- Duplicate-email check-then-write is not atomic. Fine for the mutex-guarded
  in-memory store; a real database adapter would need a unique constraint. Noted,
  not solved here.
- `GET /users/count` must not be captured by `/users/{id}`; Go 1.22+ `ServeMux`
  prefers the more specific pattern, and the test pins it.
