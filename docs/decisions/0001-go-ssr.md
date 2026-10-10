# ADR 0001: Go server-side rendering, no frontend framework

**Status:** accepted

**Context.** Pomosky's interesting client code (timer, WebSocket, sky canvas) is plain browser
JavaScript regardless of the page framework. A separate SPA would add a second deployment,
CORS and cross-domain cookies.

**Decision.** Render pages with Go `html/template` (MVC), serve them from the same server, and
add small TypeScript modules for interactivity.

**Consequences.** One deploy, same-origin cookies, Open Graph tags for free. Heavy interactive
UI later may justify htmx or a framework for specific pages.
