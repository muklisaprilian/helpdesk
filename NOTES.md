# Architecture notes

## Decisions

- **Layered Go backend**: `handlers → services → repository → PostgreSQL`. Business rules (status transitions, ticket visibility, internal notes) live in the service layer so they are unit-testable without HTTP.
- **Gin + JWT**: lightweight and familiar; middleware checks Bearer tokens and role gates for admin routes.
- **Idempotent SQL migrations**: tracked in `schema_migrations`, safe to re-run on every boot. Seed creates one account per role plus sample tickets.
- **Vue 3 Composition API + Pinia + Vue Router**: session state in Pinia, centralized Axios client for auth headers and 401 logout, route guards for auth/admin pages.
- **Docker Compose**: one command runs Postgres, API, and nginx-served frontend.

## Intentionally left out

- Attachment uploads, WebSocket/SSE notifications, SLA timers, and admin analytics (listed as optional bonuses).
- CI pipeline (can be added with GitHub Actions running `go test` and `npm run build`).
- Refresh tokens / password reset flows.

## If I had more time

- Integration tests against a testcontainers Postgres instance.
- Soft-delete / audit log for status changes.
- Stronger frontend form validation library and E2E smoke tests (Playwright).
- Rate limiting and structured request logging.
