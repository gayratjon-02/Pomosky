# ADR 0002: Pure room core (functional core, imperative shell)

**Status:** accepted

**Decision.** All room rules live in `internal/room` as `Reduce(state, event, now)`. It imports only the
standard library and has no clock, randomness or I/O. Adapters and controllers stay thin.

**Consequences.** Rules are testable with plain tables. The dependency rules are enforced by
`internal/archtest`.
