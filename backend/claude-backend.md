# Backend V1 Plan: ZenList Core Domain

## Summary
- Build `/Users/faiz/faiz/personal/zenlist/backend/go-graphql` as a Go GraphQL backend using `gqlgen`, `sqlc + pgx`, PostgreSQL, and `golang-migrate`.
- Implement first-phase entities and flows: single-user profile, projects, user-global labels, tasks, and one-level subtasks.
- Support task scheduling (`startAt`, `dueAt`), deadlines, priority (`P1`-`P5`), status (`TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`), and multi-label assignment.
- Apply soft delete for core entities and cursor pagination + status/priority filters for list queries.

## Architecture and Repository Layout
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/cmd/api/main.go`: app bootstrap, HTTP server, GraphQL endpoint, health endpoint.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/internal/config`: environment config parsing.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/internal/db`: pgx pool initialization and transaction helpers.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/internal/store`: sqlc-generated queries and thin repository wrappers.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/internal/service`: business rules/validation for profile, projects, labels, tasks.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/schema`: GraphQL SDL files.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/graph`: gqlgen generated types + resolver implementations.
- `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/migrations`: SQL migration files.

## Data Model (PostgreSQL)
1. `users`
- `id UUID PK`
- `name TEXT NOT NULL`
- `email TEXT NOT NULL UNIQUE`
- `timezone TEXT NOT NULL`
- `avatar_url TEXT NULL`
- `created_at TIMESTAMPTZ NOT NULL`
- `updated_at TIMESTAMPTZ NOT NULL`
- `deleted_at TIMESTAMPTZ NULL`

2. `projects`
- `id UUID PK`
- `user_id UUID NOT NULL FK users(id)`
- `title TEXT NOT NULL`
- `description TEXT NULL`
- `color TEXT NULL` (validate `#RRGGBB`)
- `created_at`, `updated_at`, `deleted_at`

3. `labels` (user-global)
- `id UUID PK`
- `user_id UUID NOT NULL FK users(id)`
- `name TEXT NOT NULL`
- `created_at`, `updated_at`, `deleted_at`
- unique index on `(user_id, lower(name))` where `deleted_at IS NULL`

4. `tasks`
- `id UUID PK`
- `user_id UUID NOT NULL FK users(id)`
- `project_id UUID NOT NULL FK projects(id)`
- `parent_task_id UUID NULL FK tasks(id)`
- `title TEXT NOT NULL`
- `description TEXT NULL`
- `status TEXT NOT NULL` with check in (`TODO`,`IN_PROGRESS`,`BLOCKED`,`DONE`)
- `priority TEXT NOT NULL` with check in (`P1`,`P2`,`P3`,`P4`,`P5`)
- `start_at TIMESTAMPTZ NULL`
- `due_at TIMESTAMPTZ NULL`
- `completed_at TIMESTAMPTZ NULL`
- `created_at`, `updated_at`, `deleted_at`
- check `due_at >= start_at` when both present

5. `task_labels`
- `task_id UUID NOT NULL FK tasks(id)`
- `label_id UUID NOT NULL FK labels(id)`
- primary key `(task_id, label_id)`

## GraphQL Public API (Schema-First)
### Scalars and Enums
- `scalar Time`
- `enum TaskStatus { TODO IN_PROGRESS BLOCKED DONE }`
- `enum TaskPriority { P1 P2 P3 P4 P5 }`

### Core Types
- `User`, `Project`, `Label`, `Task`
- `Task` includes `parentTaskId`, `subtasks`, `labels`, schedule/deadline/priority/status fields
- `PageInfo { endCursor, hasNextPage }`
- `ProjectConnection`, `TaskConnection`, `LabelConnection`

### Queries
- `me: User!`
- `projects(first: Int = 20, after: String): ProjectConnection!`
- `project(id: ID!): Project`
- `labels(first: Int = 50, after: String): LabelConnection!`
- `tasks(projectId: ID!, parentTaskId: ID, statuses: [TaskStatus!], priorities: [TaskPriority!], first: Int = 20, after: String): TaskConnection!`
- `task(id: ID!): Task`

### Mutations
- `upsertMe(input): User!`
- `createProject(input): Project!`
- `updateProject(input): Project!`
- `deleteProject(id): DeletePayload!`
- `createLabel(input): Label!`
- `updateLabel(input): Label!`
- `deleteLabel(id): DeletePayload!`
- `createTask(input): Task!` (single mutation for task and subtask via optional `parentTaskId`)
- `updateTask(input): Task!`
- `deleteTask(id): DeletePayload!`

### Input/Output Contracts
- `CreateTaskInput` includes `projectId`, optional `parentTaskId`, `title`, optional `description`, optional `startAt`, optional `dueAt`, optional `priority` (default `P3`), optional `status` (default `TODO`), optional `labelIds`.
- `UpdateTaskInput` supports editable fields except `parentTaskId` (kept immutable in v1 for simpler rules).
- `DeletePayload` returns `id` and `deletedAt`.

## Business Rules and Validation
- Single-user mode: no auth; every resolver scopes operations to one default user row.
- Default user is ensured at startup from env (`DEFAULT_USER_NAME`, `DEFAULT_USER_EMAIL`, `DEFAULT_USER_TIMEZONE`, `DEFAULT_USER_AVATAR_URL`).
- One-level subtask rule: `parentTaskId` can reference only a root task (`parent_task_id IS NULL`).
- Parent and subtask must belong to same `user_id` and `project_id`.
- `DONE` transition auto-sets `completedAt` to current UTC time; moving away from `DONE` clears `completedAt`.
- Soft-delete behavior:
- Deleting project sets `projects.deleted_at` and soft-deletes all tasks/subtasks in that project in one transaction.
- Deleting task soft-deletes the task and its direct subtasks in one transaction.
- Deleting label soft-deletes label and removes `task_labels` links for that label.
- All reads exclude `deleted_at IS NOT NULL`.

## Data Access and Query Strategy
- `sqlc` owns SQL in `/Users/faiz/faiz/personal/zenlist/backend/go-graphql/internal/store/sql`.
- Cursor pagination uses `(created_at, id)` tuple encoded as opaque base64 cursor.
- Default ordering:
- Projects: `created_at DESC, id DESC`
- Tasks: `COALESCE(due_at, 'infinity') ASC, priority ASC, created_at DESC, id DESC`
- Resolver layer uses batched loading for task labels and subtasks to avoid N+1 fetch patterns.

## Implementation Phases
1. Bootstrap
- Initialize Go module, config package, pgx connection, GraphQL server wiring, `/healthz`.
- Add migration runner command and local startup scripts.

2. Migrations
- Create initial schema and indexes for tables above.
- Add check constraints and partial unique indexes.

3. Store Layer (sqlc)
- Author SQL for CRUD + list/pagination/filter queries for users/projects/labels/tasks.
- Add transactional operations for delete cascades and task completion semantics.

4. Service Layer
- Centralize validation and domain rules.
- Implement status transition handling and one-level subtask enforcement.

5. GraphQL Schema + Resolvers
- Define SDL files.
- Generate gqlgen code.
- Implement resolvers by calling service layer.
- Standardize GraphQL error codes/messages for validation and not-found cases.

6. Observability and Ops
- Structured request logging.
- Basic resolver latency/error counters.
- Health check includes DB ping.

## Test Cases and Scenarios
- Migration tests: apply up/down cleanly on empty database.
- Profile tests: startup auto-creates default user; `upsertMe` updates existing row.
- Project tests: create/update/list/delete with pagination and soft-delete filtering.
- Label tests: create/update/delete; unique label name per user (case-insensitive).
- Task tests: create root task; create subtask under root; reject subtask under subtask.
- Task schedule tests: reject `dueAt < startAt`; accept null schedules.
- Status tests: transition to `DONE` sets `completedAt`; reopening clears it.
- Filter tests: list tasks filtered by status and priority with cursor paging.
- Delete tests: deleting project soft-deletes all tasks/subtasks; deleting task soft-deletes direct subtasks.
- Resolver tests: GraphQL schema contract and error responses for invalid IDs and rule violations.
- Integration scenario: end-to-end flow user -> project -> labels -> task -> subtask -> filter -> delete.

## Assumptions and Defaults
- No authentication in v1; backend is single-user mode by design.
- PostgreSQL is mandatory for v1.
- Subtasks are exactly one level deep.
- Labels are user-global and can be attached to both tasks and subtasks with many-to-many mapping.
- Task `parentTaskId` is immutable after creation in v1.
- Out of scope for this phase: sharing/collaboration, reminders/notifications, recurring tasks, comments, attachments, activity history.
