# ZenList Backend Coding Guidelines

These guidelines define how backend code must be structured and shipped for the Go + GraphQL + PostgreSQL stack (`gqlgen`, `sqlc`, `pgx`).

## 1) Repository structure and boundaries

Hard rule: keep GraphQL (transport) separate from domain and DB.

Preferred layout:

- `cmd/api/` - main entrypoint, wiring
- `internal/config/` - config parsing and validation
- `internal/graphql/`
- `schema/` - `.graphqls`
- `resolver/` - gqlgen resolvers only (thin)
- `middleware/` - auth, tracing, etc.
- `internal/service/` - business use-cases (pure-ish Go)
- `internal/db/`
- `sqlc/` - generated queries
- `repo/` - repositories wrapping sqlc
- `migrations/`
- `internal/platform/`
- `logger/`, `metrics/`, `tracing/`, `http/`
- `pkg/` only if truly needed as a reusable library outside this repo

Resolver rule:

- Resolvers should call `service.*`, not `repo.*` directly, except trivial read-only setups.

## 2) Code generation discipline (`gqlgen` + `sqlc`)

### sqlc

- Keep SQL in `internal/db/queries/*.sql` and treat it as API.
- Enforce `:one`, `:many`, `:exec`, `:execrows` usage correctly.
- Use deterministic column lists. Avoid `SELECT *` in production code.
- Always specify ordering for pagination queries.

### sqlc types

- Prefer `pgx/v5` types for scanning and `pgtype` when nullability matters.
- Map DB enums to Go enums (string type + consts).
- Validate enums at boundaries.

### gqlgen

- Keep GraphQL schema under version control as the public contract.
- Use gqlgen config to map scalars (UUID, Time) to strong Go types.
- Do not leak DB models into GraphQL models.

### generated code drift

- Regenerate in CI and fail if git diff appears:
- `go generate ./...` (or explicit `sqlc generate` + `gqlgen generate`)
- This prevents generated code drift.

## 3) Database access rules (`pgx` + `sqlc`)

### connection pooling

- Use `pgxpool` and tune with real load.
- Configure `MaxConns` based on CPU and DB limits.
- Configure `MinConns` for steady latency.
- Set `HealthCheckPeriod`.
- Set query timeouts via `context.WithTimeout` at the service boundary.

### transactions

- Make transaction boundaries explicit in the service layer.
- Use a repo pattern that accepts a `Querier` (`sqlc` interface) bound to either:
- pool (default)
- tx (transaction)
- Rule: no resolver starts a transaction; only services do.

### query correctness

- Every list endpoint uses stable ordering and cursor pagination.
- All multi-tenant queries must include `tenant_id = $1` in SQL (never filtered in Go).
- Use DB constraints (`UNIQUE`, `FK`, `CHECK`) and treat violations as expected errors.

## 4) GraphQL production patterns

### N+1 prevention

- Use dataloaders for any field resolver that could fan out.
- Batch per request and cache per request.
- Fetch by IDs in one SQL query.

### resolver behavior

- Resolvers must be deterministic.
- Resolvers should keep side effects minimal.
- Reads must be idempotent.

### error model

- Never return raw DB errors to clients.
- Convert errors to GraphQL errors with:
- user-safe message
- machine-readable `extensions.code`
- request id for support

Suggested categories:

- `UNAUTHENTICATED`, `FORBIDDEN`
- `BAD_USER_INPUT`
- `NOT_FOUND`, `CONFLICT`
- `INTERNAL`

### complexity control

- Enforce max query depth, complexity, or cost.
- Limit list sizes with defaults and hard maximums.
- Disable introspection in public production where relevant.

## 5) Context, cancellation, timeouts

- Every DB call must use request context (never `context.Background()`).
- Set timeouts by tier:
- HTTP request: 10-30s
- DB query: 1-3s typical, 5-10s heavy endpoints
- background jobs: explicit longer timeouts
- Propagate via typed context keys:
- auth claims
- request id / trace id
- tenant id

## 6) Logging, metrics, tracing (operability)

### logging

- Structured JSON logs only.
- Log fields: `request_id`, `trace_id`, `user_id`, `tenant_id`, `operationName`, `duration_ms`.
- Do not log PII or raw tokens.
- Log errors once at boundary, not in every layer.

### metrics

- Request count and latency histogram by operation name.
- DB query latency and error counts.
- Pool stats (`AcquireCount`, `AcquireDuration`, `MaxConns`, etc.).

### tracing

- One trace per GraphQL request.
- Child spans: resolvers (optional, sampled), DB calls (required).

## 7) AuthZ and multi-tenancy

- AuthN at middleware (extract identity).
- AuthZ in service methods ("can user do X?").
- Never trust client-provided IDs for tenant scoping.
- Keep RBAC/ABAC rules in service layer, not resolvers.

## 8) Validation and input hygiene

- Validate inputs at GraphQL boundary:
- length
- format
- enums
- required fields
- Validate business invariants in service layer.
- DB constraints are the final line of defense.
- Fail fast with `BAD_USER_INPUT`.

## 9) Testing strategy

### unit tests

- Service layer tests with fake repos (interfaces).
- Table-driven tests with edge cases.

### DB integration tests

- Run PostgreSQL in tests (container).
- Apply migrations, then test repositories and SQL correctness.
- Avoid mocking `sqlc`; it hides SQL mistakes.

### GraphQL tests

- Maintain golden tests for major operations:
- auth behavior
- pagination behavior
- error codes

## 10) Migrations and schema ownership

- Use a real migration tool (`goose`, `atlas`, or `migrate`).
- Every PR with schema changes includes:
- migration up/down (if feasible)
- updated `sqlc` queries if needed
- backfill plan if required
- Never run destructive migrations without rollout plan:
- expand -> backfill -> switch reads -> contract

## 11) Concurrency and performance rules

- Do not parallelize DB calls blindly; use dataloader batching first.
- If concurrency is needed:
- cap with semaphore or `errgroup`
- propagate context
- avoid shared mutable state without locks
- Cache policy:
- per-request dataloader cache
- optional cross-request cache only for stable data (feature flags, config)

## 12) CI/CD quality gates

Minimum gates before merge:

- `go test ./...`
- `go vet ./...`
- `golangci-lint run`
- `sqlc generate` + `gqlgen generate` with clean git diff
- `staticcheck`
- `go test -race` (nightly if too slow for every PR)

## 13) No foot-guns checklist

- No `SELECT *`.
- No DB calls inside resolver loops unless via dataloader.
- No business logic or transaction start inside resolvers.
- No raw internal errors returned to clients.
- No missing tenant filters in SQL.
- No pagination without stable ordering.
- No request path without timeout.
- No PII or token logging.
