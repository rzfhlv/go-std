# Requirement: Complete the user module's feature surface

- Status: done
- Date: 2026-10-20

## Problem

The user module is the repo's minimal reference slice, but it only exposes a
partial lifecycle. Operators can create, fetch, list, and (in the CLI) count
users, but they cannot change a user's name or remove one. A rename is already
modelled for an individual user but no request reaches it, so the capability
exists on paper while being unusable. In addition, two behaviours are
inconsistent: the CLI can count users but the REST API cannot, and email
uniqueness is not enforced, so the same person can be stored twice under
different ids.

This makes the module a weak reference: contributors copying `user` for a new
bounded context inherit an incomplete, inconsistent feature set.

## Solution

The user module gains the missing parts of the lifecycle, exposed consistently
across every way the module is used.

An operator can rename an existing user and delete a user. Renaming and deleting
an unknown user fails with a clear "not found" outcome; renaming to an empty name
fails with a clear "invalid" outcome. Deleting a user removes it from every
subsequent read.

Rejecting a duplicate: creating a second user with an email already in use fails
with a clear, distinct "email already exists" outcome, so callers can tell it
apart from malformed input.

Both the REST API and the CLI expose the same capability set: create, fetch one,
list, count, rename, and delete. Each channel reports the same failures the
same way its existing errors are reported (HTTP status codes for REST, non-zero
exit with a message for CLI).

## Acceptance criteria

Mark a criterion done by replacing its `- [ ]` with `- ✅ ` (checkmark) once the
behavior is implemented and verified; leave it `- [ ]` until then.

- ✅ Given an existing user, when the operator renames them to a non-empty
      name, then subsequent reads return the new name and the same id and email.
- ✅ Given a user whose name is empty or whitespace, when renamed, then the
      operation fails as invalid and the stored name is unchanged.
- ✅ Given a nonexistent user id, when renamed, then the operation fails as
      not found.
- ✅ Given an existing user, when deleted, then fetching or listing no longer
      returns them, and deleting them again fails as not found.
- ✅ Given a user already stored with an email, when a new user is created with
      the same email (regardless of case or surrounding whitespace), then
      creation fails as a duplicate and no second user is stored.
- ✅ Each of create, get, list, count, rename, and delete is reachable over
      both REST and the CLI, and both report the invalid / not-found / duplicate
      failures consistently with their existing error handling.
- ✅ All existing tests keep passing; every new behavior has table-driven tests
      covering the happy path and the error path.

## Out of scope

- gRPC implementation (remains a documented placeholder until protobuf and the
  `google.golang.org/grpc` dependency land).
- Authentication, authorization, or ownership of users.
- Pagination, sorting, or filtering of the user list.
- Persistence beyond the in-memory store (no database-backed storage here).
- Bulk operations (bulk delete, bulk create).

## Notes

- The module already models a rename at the level of an individual user that no
  request currently reaches; this requirement should make it reachable rather
  than restate the rule.
- Email uniqueness is a new invariant. It has to be decided against the full set
  of stored users rather than one user in isolation, so the rule "an email is
  well-formed" stays with the individual user, while "this email is already
  taken" is a separate, externally visible outcome expressed as its own error so
  callers can match on it.
- The transient in-memory store is one-shot per process, so acceptance is judged
  within a single process run (seed, act, read back), matching how the existing
  tests exercise the module.
- Completing this makes `user` a faithful small reference for the `order` module's
  richer lifecycle, and it is the natural first candidate for a plan written with
  `/plan complete-user-module`.
