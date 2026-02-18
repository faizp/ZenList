# ZenList

ZenList is an ultra-minimal todo list app built with:

- React frontend
- Go + GraphQL backend
- PostgreSQL

The frontend includes a minimal API tester UI to exercise:

- User profile (`me`, `upsertMe`)
- Projects (create, update, delete, list)
- Labels (create, update, delete, list)
- Tasks and subtasks (create, update, delete, list)
- Status/priority filtering and label assignment

## Run with Docker (recommended)

### Prerequisites

- Docker Desktop (or Docker Engine) running

### Start everything

From project root:

```bash
docker compose up --build
```

Services:

- Frontend: `http://localhost:5173`
- Backend GraphQL endpoint: `http://localhost:8080/query`
- Backend Playground: `http://localhost:8080/`
- Backend Health: `http://localhost:8080/healthz`
- PostgreSQL: `localhost:5432` (`zenlist` / `zenlist`)

### Stop services

```bash
docker compose down
```

To also remove DB data volume:

```bash
docker compose down -v
```

## Run locally (without full Docker stack)

This mode runs PostgreSQL in Docker and backend/frontend directly on your machine.

### 1) Start PostgreSQL only

```bash
docker compose up -d postgres
```

### 2) Run backend

```bash
cd backend/go-graphql
cp .env.example .env
make migrate-up
make run
```

### 3) Run frontend

In a new terminal:

```bash
cd apps/web-react
npm install
npm run dev
```

Then open `http://localhost:5173`.

## Useful commands

### Backend tests

```bash
cd backend/go-graphql
make test
```

### Backend codegen

```bash
cd backend/go-graphql
make gen
```
