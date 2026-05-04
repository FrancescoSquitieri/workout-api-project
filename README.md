# Workout API — Go Backend

An HTTP backend for managing **workouts** with users, **opaque token** authentication, and **PostgreSQL** persistence. I built this while **learning Go from the ground up**; it doubles as **my portfolio** backend sample. It showcases idiomatic libraries, layered architecture, versioned migrations, and integration-style tests against the data layer.

## What the application does

The service exposes a **JSON REST-like API** that lets you:

1. Register a user (`POST /users`) with username, email, password (**bcrypt** hash), and optional bio.
2. Obtain an **authentication token** (`POST /tokens/auth`) after credential checks; the token lifetime is fixed in code (**24 hours**). Only the **SHA-256 hash** is stored in the database—the plaintext value is returned once to the client.
3. Manage **workouts** with the `Authorization: Bearer <token>` header:
   - create with nested **entries** (exercises),
   - read by ID,
   - partial updates (merged fields and entries),
   - delete.  
   Update/delete paths ensure the workout belongs to the **current user**.

There is also **`GET /health`** for a simple application smoke check.

### Data model (high level)

- **User**: credentials and profile fields.
- **Workout**: title, description, duration, calories; linked via `user_id`.
- **Workout entries**: sets, reps *or* duration (per SQL constraints), optional weight, notes, ordering.
- **Token**: hash, `user_id`, expiry, scope (`authentication`).

## Tech stack

| Area | Choice in this project |
|------|-------------------------|
| Language | **Go** (module `apiProject`) |
| HTTP | **`net/http`** + [**chi**](https://github.com/go-chi/chi) v5 router |
| Database | **PostgreSQL** via **`database/sql`** and [**pgx**](https://github.com/jackc/pgx) (stdlib) |
| Migrations | [**goose**](https://github.com/pressly/goose) v3, SQL **embedded** in the binary with **`embed`** |
| Passwords | [**bcrypt**](https://pkg.go.dev/golang.org/x/crypto/bcrypt) (`golang.org/x/crypto`) |
| Tests | **`testing`** + [**testify**](https://github.com/stretchr/testify) |

## Repository layout

```
.
├── main.go                 # Entry: port flag, HTTP server, timeouts
├── docker-compose.yml      # Dev Postgres + test Postgres (host ports 5432 and 5433)
├── migrations/             # goose migrations + embedded filesystem
│   ├── fs.go               //go:embed *.sql
│   ├── 00001_users.sql
│   ├── 00002_workouts.sql
│   ├── 00003_workout_entires.sql   # filename typo (should be "entries")
│   ├── 00004_tokens.sql
│   └── 00005_user_id_alter.sql     # Adds user_id to workouts
├── internal/
│   ├── app/app.go          # App wiring: DB, migrate, handlers, middleware
│   ├── routes/routes.go    # Route registration and protected groups
│   ├── middleware/         # Authenticate (Bearer), RequireUser, user context
│   ├── api/                  # HTTP handlers: user, token, workout
│   ├── store/                # PostgreSQL access + interfaces
│   ├── tokens/               # Random tokens and scope constants
│   └── utils/                # JSON envelope, chi URL param `id`
└── database/                 # Container volume data (gitignored as needed)
```

## Internal flow (high level)

1. **`main.go`** parses the port (`-port`, default **3005**), builds `Application`, and starts `http.Server` with read/write timeouts.
2. **`app.NewApplication`** opens the DB, runs **`MigrateFS`** over the `migrations` package (schema aligned on startup), then constructs stores, handlers, and middleware.
3. **`Authenticate`** middleware sets `Vary: Authorization`, validates the Bearer token by hashing and joining `tokens`, and injects the user into **`context`** (or an anonymous user if the header is missing).
4. **`RequireUser`** returns 401 when the request is unauthenticated; workout handlers additionally verify **ownership** for update/delete.

## SQL migrations (summary)

| File | Purpose |
|------|---------|
| `00001_users.sql` | `users` table with unique username/email |
| `00002_workouts.sql` | `workouts` table (initially without `user_id`) |
| `00003_workout_entires.sql` | `workout_entries` with FK and CHECK on reps vs `duration_seconds` |
| `00004_tokens.sql` | `tokens` table (hash PK, FK to `users`) |
| `00005_user_id_alter.sql` | `ALTER TABLE workouts ADD COLUMN user_id ...` |

Migrations run automatically at app startup via **goose Up** on the embedded filesystem.

## Running locally

### Prerequisites

- [Go](https://go.dev/dl/) matching the toolchain in `go.mod`
- Docker (optional but recommended) for PostgreSQL

### Database with Docker Compose

```bash
docker compose up -d
```

This starts two services:

- **`db`**: database `workoutDB` on host port **5432**
- **`test_db`**: database `workoutDB_test` on host port **5433** (used by tests)

Credentials in code (`internal/store/database.go`) match Compose: user/password `postgres` / `postgres`.

### Run the API

```bash
go run . -port 3005
```

Logs show the listening port; migrations run during bootstrap.

## API endpoints (quick reference)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/health` | No | Application heartbeat |
| `POST` | `/users` | No | Register user (JSON) |
| `POST` | `/tokens/auth` | No | Login; JSON includes `auth_token` with plaintext token and `expiry` |
| `GET` | `/workouts/{id}` | Bearer | Workout detail + entries |
| `POST` | `/workouts` | Bearer | Create workout (`entries` in body) |
| `PUT` | `/workouts/{id}` | Bearer | Update (owner only) |
| `DELETE` | `/workouts/{id}` | Bearer | Delete (owner only) |

Example header: `Authorization: Bearer <plaintext_from_tokens_auth>`.

## Tests

Tests in `internal/store/workout_store_test.go` expect PostgreSQL on **port 5433** and goose migrations from `../../migrations`. With `test_db` running:

```bash
go test ./internal/store/... -v
```

They cover successful workout creation and **constraint violations** on entries (e.g. reps and duration set at once).

## Why this project

What I aimed to practice with my **first REST API in Go**:

- **`handler → store → SQL`** boundaries via interfaces (`WorkoutStore`, `UserStore`, `TokenStore`).
- **Transactions** for multi-row workout create/update (`workouts` + `workout_entries`).
- **Middleware** scoped to route groups for auth and authorization.
- **Deterministic migrations** embedded and applied on each local bootstrap.

**Portfolio one-liner:** *Go REST API on PostgreSQL, hashed opaque tokens, goose migrations, chi router, and data-layer integration tests.*

---

Maintained as **my** reference implementation—hands-on experiment and something I’m happy to walk through in interviews or demos.
# workout-api-project
