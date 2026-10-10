# ADR 0004: Login is optional

**Status:** accepted

**Decision.** Anonymous visitors can use every feature. An account adds statistics, a saved nickname
and room ownership. `OptionalAuth` never blocks a request; only host actions are protected (host token
or room owner).

**Consequences.** No login wall for viral traffic. Passwordless login (Google OAuth or magic link) avoids
storing passwords.
