ThreadTalk is a small forum-style backend API written in Go. It provides user
authentication (signup/login), topics, posts and threaded comments with basic
pagination and JWT-based auth.

**Tech stack:** Go (Gin), PostgreSQL, golang-migrate, Docker

## Quickstart

Prerequisites: `Go 1.25+`, `Docker` (optional for containerized run), and a
PostgreSQL instance.

1. Copy or create a `.env` file in the project root (see the provided `.env` for
	 a working example). Make sure `DATABASE_URL` and `PORT` are set.
2. To run with Docker (recommended):

```bash
docker-compose up --build
```

3. To run locally without Docker:

```bash
# Start Postgres locally and ensure DATABASE_URL points to it
go run ./cmd/api
```

The server will run on `http://localhost:${PORT}` (default `8080`). On startup,
database migrations in `internal/db/migrations` are applied automatically.

## Configuration
Key environment variables (see `.env`):

- `DB_USER`, `DB_PASSWORD`, `DB_NAME` — used by `docker-compose` when running
	Postgres
- `DATABASE_URL` — full Postgres connection string used by the app
- `JWT_SECRET` — used to sign authentication tokens
- `FRONTEND_URL`, `BACKEND_URL` — CORS / allowed origins
- `PORT` — HTTP server port (default `8080`)

## Project structure

- `cmd/api` — application entry (`main.go`)
- `internal/db` — DB connection and migrations
- `internal/handlers` — HTTP handlers (auth, forum, etc.)
- `internal/middleware` — middleware (auth)
- `internal/models` — domain models (user, forum/topic)
- `internal/router` — router setup
- `internal/utils` — helpers (JWT helpers)
- `internal/db/migrations` — SQL migration files
- `API.md` — API endpoint summaries and examples

## API
See [API.md](API.md) for a short summary of available endpoints (signup, login,
topics, posts, comments, profile).

## Development notes

- The Dockerfile builds a static Go binary and copies migrations into the
	image, so the container runs migrations at startup.
- Use `go env` and `go run ./cmd/api` to run the server locally.

If you'd like, I can add a `.env.example` and a short health-check endpoint next.
golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/postgres
go get github.com/golang-migrate/migrate/v4/source/file
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/gin-contrib/cors