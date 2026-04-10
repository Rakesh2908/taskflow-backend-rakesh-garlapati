## Overview

`taskflow` is a Go REST API for basic task management: users can register/login, create projects, and create/update/list/delete tasks within projects.

- **Tech stack**:
  - **Go** (HTTP server + business logic)
  - **chi** router (`github.com/go-chi/chi/v5`)
  - **PostgreSQL** (data store)
  - **pgx/v5** (`github.com/jackc/pgx/v5/pgxpool`) for DB access
  - **golang-migrate** (SQL migrations)
  - **JWT** (`github.com/golang-jwt/jwt/v5`) for auth
  - **bcrypt** (`golang.org/x/crypto/bcrypt`) for password hashing
  - **validator** (`github.com/go-playground/validator/v10`) for request validation

## Architecture Decisions

- **Layering**
  - **`repository/`**: interfaces + pg-backed implementations (testable; hides `pgx` errors behind `utils` sentinels).
  - **`service/`**: business rules (auth, permissions, partial updates). Depends only on repository interfaces + config + models/utils.
  - **`api/handler/`**: HTTP-only concerns (decode/validate, error mapping, status codes). Depends on service interfaces, not implementations.
  - **`middleware/`**: pure HTTP middleware (JWT auth + request logger).
  - **`migrations/`**: raw SQL migrations (no ORM).

- **Tradeoffs**
  - **Minimal validation semantics**: request validation uses `validator/v10` tags and returns a simple `{fields: {field: tag}}` map.
  - **Simple error model**: service/repo return `utils` sentinels (`ErrNotFound`, `ErrForbidden`, `ErrConflict`, etc.) mapped in handlers.
  - **JWT claims**: token includes `user_id` and `email` claims to keep middleware stateless (no DB calls in auth middleware).

- **Intentionally left out**
  - **No refresh tokens / sessions** (kept auth simple for an assignment).
  - **No pagination / sorting options** on list endpoints.
  - **No OpenAPI/Swagger generation** (API reference is documented below).

## Running Locally

Assuming you have **Docker** installed (and nothing else):

```bash
git clone https://github.com/Rakesh2908/taskflow
cd taskflow
cp .env.example .env
docker compose up --build
```

- **API base URL**: `http://localhost:8080`

## Environment Variables (.env)

Create a `.env` file in the repo root (do **not** commit it). You can copy the template:

```bash
cp .env.example .env
```

These variables are used by Docker Compose and the API:

```env
# API
JWT_SECRET=change-me
PORT=8080
BCRYPT_COST=12

# Postgres (used by docker compose postgres)
POSTGRES_DB=taskflow
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

# Database connection string used by the API + migrations
# (inside docker-compose, host is `postgres`)
DB_URL=postgres://postgres:postgres@postgres:5432/taskflow?sslmode=disable
```

## Backend Quickstart (end-to-end)

This is the quickest way to start the backend and hit every endpoint.

```bash
# 1) Start the backend + database
cp .env.example .env
docker compose up --build
```

In another terminal:

```bash
BASE_URL="http://localhost:8080"

# 2) Login with seeded user (no registration needed)
TOKEN="$(curl -sS -X POST "$BASE_URL/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}' | \
  python -c 'import sys,json; print(json.load(sys.stdin)["token"])')"

echo "TOKEN=$TOKEN"

# 3) Create a project
PROJECT_ID="$(curl -sS -X POST "$BASE_URL/projects" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Demo","description":"My project"}' | \
  python -c 'import sys,json; print(json.load(sys.stdin)["id"])')"

echo "PROJECT_ID=$PROJECT_ID"

# 4) Create a task in the project
TASK_ID="$(curl -sS -X POST "$BASE_URL/projects/$PROJECT_ID/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Ship it","description":"...","priority":"low"}' | \
  python -c 'import sys,json; print(json.load(sys.stdin)["id"])')"

echo "TASK_ID=$TASK_ID"

# 5) List projects
curl -sS "$BASE_URL/projects" -H "Authorization: Bearer $TOKEN"

# 6) Get project by id
curl -sS "$BASE_URL/projects/$PROJECT_ID" -H "Authorization: Bearer $TOKEN"

# 7) List tasks (optionally filter by status)
curl -sS "$BASE_URL/projects/$PROJECT_ID/tasks" -H "Authorization: Bearer $TOKEN"
curl -sS "$BASE_URL/projects/$PROJECT_ID/tasks?status=todo" -H "Authorization: Bearer $TOKEN"

# 8) Update task (PATCH)
curl -sS -X PATCH "$BASE_URL/tasks/$TASK_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"done"}'

# 9) Delete task
curl -sS -X DELETE "$BASE_URL/tasks/$TASK_ID" -H "Authorization: Bearer $TOKEN" -i

# 10) Delete project
curl -sS -X DELETE "$BASE_URL/projects/$PROJECT_ID" -H "Authorization: Bearer $TOKEN" -i
```

## Running Migrations

Migrations are run **automatically on API container startup** via `entrypoint.sh`:

```sh
migrate -path /migrations -database "$DB_URL" up
exec /server
```

If you want to run them manually (inside the API container):

```bash
docker compose exec api migrate -path /migrations -database "$DB_URL" up
```

## Test Credentials

Seeded user (from `migrations/000002_seed.up.sql`):

- **Email**: `test@example.com`
- **Password**: `password123`

## API Reference

### Auth (public)

#### POST `/auth/register`

Request:

```bash
curl -sS -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com","password":"password123"}'
```

Response: `201`

```json
{
  "token": "JWT_HERE",
  "user": { "id": "UUID", "name": "Alice", "email": "alice@example.com" }
}
```

#### POST `/auth/login`

```bash
curl -sS -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}'
```

Response: `200`

```json
{
  "token": "JWT_HERE",
  "user": { "id": "UUID", "name": "Test User", "email": "test@example.com" }
}
```

### Authenticated requests

All routes below require:

```bash
-H "Authorization: Bearer $TOKEN"
```

### Projects

#### GET `/projects`

```bash
curl -sS http://localhost:8080/projects \
  -H "Authorization: Bearer $TOKEN"
```

Response: `200` (array of projects)

#### POST `/projects`

```bash
curl -sS -X POST http://localhost:8080/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Demo","description":"My project"}'
```

Response: `201`

#### GET `/projects/:id`

```bash
curl -sS http://localhost:8080/projects/$PROJECT_ID \
  -H "Authorization: Bearer $TOKEN"
```

Response: `200`

#### PATCH `/projects/:id`

```bash
curl -sS -X PATCH http://localhost:8080/projects/$PROJECT_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Renamed"}'
```

Response: `200`

#### DELETE `/projects/:id`

```bash
curl -sS -X DELETE http://localhost:8080/projects/$PROJECT_ID \
  -H "Authorization: Bearer $TOKEN" -i
```

Response: `204`

### Tasks

#### GET `/projects/:id/tasks`

Optional query params:
- `status` = `todo|in_progress|done`
- `assignee` = UUID

```bash
curl -sS "http://localhost:8080/projects/$PROJECT_ID/tasks?status=done" \
  -H "Authorization: Bearer $TOKEN"
```

Response: `200` (array of tasks)

#### POST `/projects/:id/tasks`

```bash
curl -sS -X POST http://localhost:8080/projects/$PROJECT_ID/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Ship it","description":"...","priority":"low"}'
```

Response: `201`

#### PATCH `/tasks/:id`

```bash
curl -sS -X PATCH http://localhost:8080/tasks/$TASK_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"done"}'
```

Response: `200`

#### DELETE `/tasks/:id`

```bash
curl -sS -X DELETE http://localhost:8080/tasks/$TASK_ID \
  -H "Authorization: Bearer $TOKEN" -i
```

Response: `204`

### Error format

Non-validation errors:

```json
{ "error": "not found" }
```

Validation errors:

```json
{
  "error": "validation failed",
  "fields": { "email": "is required" }
}
```

## What You'd Do With More Time

- Add **OpenAPI** spec + generated client collection (Postman/Bruno) committed to the repo.
- Improve validation error payloads further (more specific messages per rule, localization-ready, consistent formatting across all endpoints).
- Add refresh tokens: short-lived access tokens (~15 min) + long-lived refresh tokens stored in **httpOnly cookies** (rotation + reuse detection).
- Add rate limiting: per-IP and per-user limits on auth endpoints using `golang.org/x/time/rate`.
- Add request IDs: middleware that generates a UUID per request, injects into context/logs, and returns `X-Request-ID`.
- Add stronger auth hardening: account lockout / exponential backoff, password policy, and audit logging.
- Add CI/CD: GitHub Actions pipeline (lint via `golangci-lint` → test → build Docker image → push to registry).
- Add observability: Prometheus `/metrics`, structured JSON logs with correlation IDs, and basic tracing.
- Add input sanitization: strip/escape HTML from user-provided text fields before persisting.
- Add more tests: repo-level tests with a DB, deeper service unit tests for permissions, and handler tests for error mapping.
