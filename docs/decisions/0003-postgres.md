# ADR 0003: PostgreSQL, no Redis in the MVP

**Status:** accepted

**Decision.** Persist accounts, sessions, rooms and finished focus sessions in PostgreSQL (pgx + sqlc + goose).
Live room state is in memory on a single instance, so Redis is not needed yet.

**Revisit when** more than one app instance runs (pub/sub between instances) or distributed rate limiting is needed.
