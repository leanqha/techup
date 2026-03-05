# TechUp API

Production-oriented backend service for the TechUp university application.

## What This Service Provides

- Account lifecycle: register, login, refresh, logout, profile update, password change
- Password reset flow with pluggable delivery (SMTP in production, logs in development)
- Schedule domain APIs: faculties, groups, lessons, teachers, classrooms, search
- Lesson notes for authenticated users
- Campus map APIs: buildings, rooms, shortest path between rooms
- Admin APIs for account role management and CRUD operations across schedule and map entities
- Swagger/OpenAPI docs served from the running app

## Tech Stack

- Go `1.25`
- Gin (HTTP routing)
- PostgreSQL + `pgx` connection pool
- Goose (SQL migrations)
- JWT-based auth (access + refresh token cookies)
- Zerolog (structured logging)
- Docker / Docker Compose

## Repository Layout

```text
cmd/server/          App entrypoint
internal/account/    Auth, account, password reset logic
internal/schedule/   Schedule and notes logic
internal/map/        Map, rooms, pathfinding logic
internal/health/     Health endpoint
config/              Env and database configuration
migrations/          Goose SQL migrations
docs/                Generated Swagger artifacts
```

## Prerequisites

- Go `1.25+`
- PostgreSQL `13+` (or compatible)
- Docker + Docker Compose (optional, for containerized runs)
- Goose migration CLI (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Configuration

Copy the template and fill real values:

```bash
cp .env.example .env
```

The app reads `.env` by default. You can point to another file with `ENV_FILE=/path/to/file`.

## Quickstart (5 Minutes)

```bash
cp .env.example .env
make migrate-up
go run ./cmd/server
```

Verify the service is up:

```bash
curl -i http://localhost:3000/api/v1/health
```

### Environment Variables

| Variable                     | Required              | Default | Notes                                               |
|------------------------------|-----------------------|---------|-----------------------------------------------------|
| `PORT`                       | No                    | `3000`  | HTTP port used by the app                           |
| `GIN_MODE`                   | Recommended           | `debug` | Use `release` in production                         |
| `APP_BASE_URL`               | Yes (for reset links) | -       | Public base URL used in password reset emails       |
| `DB_HOST`                    | Yes                   | -       | PostgreSQL host                                     |
| `DB_PORT`                    | Yes                   | -       | PostgreSQL port                                     |
| `DB_USER`                    | Yes                   | -       | PostgreSQL user                                     |
| `DB_PASSWORD`                | Yes                   | -       | PostgreSQL password                                 |
| `DB_NAME`                    | Yes                   | -       | PostgreSQL database name                            |
| `DB_SSLMODE`                 | Yes                   | -       | Example: `disable`, `require`, `verify-full`        |
| `JWT_SECRET`                 | Yes                   | -       | Strong random secret for token signing              |
| `JWT_ACCESS_TOKEN_TTL`       | No                    | `15`    | Minutes                                             |
| `JWT_REFRESH_TOKEN_TTL`      | No                    | `10080` | Minutes (7 days)                                    |
| `PASSWORD_RESET_TTL_MINUTES` | No                    | `30`    | Reset token lifetime in minutes                     |
| `SMTP_HOST`                  | Optional              | empty   | If empty, reset links are logged instead of emailed |
| `SMTP_PORT`                  | No                    | `587`   | SMTP port                                           |
| `SMTP_USER`                  | Optional              | empty   | SMTP username                                       |
| `SMTP_PASS`                  | Optional              | empty   | SMTP password                                       |
| `SMTP_FROM`                  | Optional              | empty   | Sender email                                        |
| `SMTP_USE_TLS`               | No                    | `false` | TLS enabled/disabled                                |
| `SMTP_SKIP_VERIFY`           | No                    | `false` | Skip TLS verification (avoid in production)         |

## Configuration Profiles

Use separate env files and switch with `ENV_FILE`:

```bash
# Development
ENV_FILE=.env go run ./cmd/server

# Production-like local run
ENV_FILE=.env.prod go run ./cmd/server

# Migrations against a specific environment file
ENV_FILE=.env.prod make migrate-up
```

Recommended profile conventions:

- `.env`: local development defaults
- `.env.prod`: production/staging values (never commit secrets)
- `.env.test`: isolated test database values if needed

## Running Locally (Go)

1. Configure environment in `.env`
2. Run database migrations
3. Start the server

```bash
make migrate-up
go run ./cmd/server
```

Service endpoints:

- API base: `http://localhost:3000/api/v1`
- Health: `http://localhost:3000/api/v1/health`
- Swagger UI: `http://localhost:3000/swagger/index.html`

## Running with Docker Compose

`docker-compose.yml` currently runs only the app container. PostgreSQL must be available externally and reachable using `DB_*` values.

```bash
make build
make up
```

Or directly:

```bash
ENV_FILE=.env docker compose -f docker-compose.yml up -d --build
```

## Database Migrations

Apply, rollback, and inspect migration state:

```bash
make migrate-up
make migrate-down
make migrate-status
make migrate-version
```

The migration directory defaults to `./migrations` and can be overridden via `GOOSE_MIGRATION_DIR`.

### Schema Notes

- Core schedule schema is created by ordered migrations in `migrations/`.
- Later migrations add lesson type support, password reset tokens, subjects, and lesson notes.
- Run `make migrate-up` before starting the app in new environments to avoid runtime query failures.
- For rollback during incidents, use `make migrate-down` cautiously and verify data impact first.

## Tests

Run all tests:

```bash
make test
```

Run tests for one module (examples):

```bash
make test-account
make test-schedule
make test-map
```

## Troubleshooting

- `failed to connect to database`: verify `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, and `DB_SSLMODE`.
- App starts but API is unreachable: confirm `PORT` in env and Docker port mapping (`${PORT}:3000`).
- Password reset emails not sent: if `SMTP_HOST` or `SMTP_FROM` is empty, reset links are logged by design.
- Browser CORS errors: backend currently allows only `https://mytechup.ru`; update CORS policy for your frontend origin.
- Migration command fails: ensure Goose is installed and `ENV_FILE` points to the expected env file.

## API Documentation (Swagger)

Swagger artifacts in `docs/` are generated from code annotations.

```bash
make swagger
```

After startup, Swagger UI is available at `/swagger/index.html`.

## Production Deployment

This repository includes CI/CD workflow: `.github/workflows/deploy-prod.yml`.

Current production flow:

1. Trigger on `main` push or manual dispatch.
2. Connect to the target host over SSH.
3. Sync repository to target branch.
4. Generate `.env.prod` from GitHub environment secrets.
5. Run `docker compose up -d --build` with `ENV_FILE=.env.prod`.

### Required GitHub Environment Secrets

- Repo/host: `PROD_HOST`, `PROD_PORT`, `PROD_USER`, `PROD_SSH_KEY`, `PROD_APP_DIR`, `PROD_REPO_URL`, `PROD_REPO_BRANCH`
- App/runtime: `PORT`, `GIN_MODE`, `APP_BASE_URL`
- Database: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- Auth: `JWT_SECRET`, `JWT_ACCESS_TOKEN_TTL`, `JWT_REFRESH_TOKEN_TTL`
- Password reset: `PASSWORD_RESET_TTL_MINUTES`
- SMTP: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`, `SMTP_USE_TLS`, `SMTP_SKIP_VERIFY`

## Operational Notes

- CORS currently allows only `https://mytechup.ru`; adjust if your frontend origin differs.
- `/api/v1/health` is unauthenticated and intended for probes.
- Keep `JWT_SECRET`, DB credentials, and SMTP credentials out of git; inject via environment.
- Use secure cookie and TLS settings at the ingress/reverse-proxy layer in production.

## Contributing

- Create a feature branch from `main`.
- Run tests locally before opening a PR:

```bash
make test
```

- Regenerate Swagger docs if handler annotations changed:

```bash
make swagger
```

- In PR description, include: scope, migration impact, config changes, and test evidence.

## License

MIT - see `LICENSE`.
