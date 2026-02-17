# ZenList Go GraphQL Backend

## Prerequisites

- Go 1.24+
- Docker (for local PostgreSQL)
- `sqlc` CLI (`sqlc generate`)

## Quick Start

1. Start PostgreSQL:

```bash
make db-up
```

2. Copy env file:

```bash
cp .env.example .env
```

3. Run migrations:

```bash
make migrate-up
```

4. Start API:

```bash
make run
```

GraphQL endpoint: `http://localhost:8080/query`
Playground: `http://localhost:8080/`
Health: `http://localhost:8080/healthz`

## Generate Code

```bash
make gen
```

## Test

```bash
make test
```
