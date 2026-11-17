# Do Together

**Do Together** is a full-stack collaborative project management application built with Go, React, PostgreSQL, and Redis.

It allows users to create projects, manage team members and roles, work with tasks and invitations, and securely manage authentication sessions.

## Features

- User registration and authentication
- Access and refresh token authentication
- Atomic refresh token rotation with Redis
- Project creation, archiving, and restoration
- Project filtering and pagination
- Role-based project access
- Project member management
- Project invitations with expiration
- Editable display names and durable JPEG/PNG avatars
- Task creation, editing, deletion, and status management
- Graceful HTTP server shutdown
- PostgreSQL and Redis health checks
- Docker-based local environment
- CI with GitHub Actions

## Tech Stack

| Technology | Purpose |
|---|---|
| **Go** | Backend application |
| **PostgreSQL** | Persistent data storage |
| **Redis** | Refresh sessions and atomic token rotation |
| **JWT** | Access token authentication |
| **REST API** | Client-server communication |
| **golang-migrate** | Database migrations |
| **Docker** | Containerization |
| **Docker Compose** | Local infrastructure |
| **GitHub Actions** | CI, tests, vetting, and builds |

---

## Architecture

Do Together is implemented as a **layered monolith**.

The application is divided into:

- HTTP handlers
- request and response DTOs
- services
- domain models
- repositories
- PostgreSQL and Redis infrastructure

### Request Flow

```text
HTTP Request
     │
     ▼
Router / Middleware
     │
     ▼
Handler
     │
     ▼
Request DTO
     │
     ▼
Service
     │
     ▼
Domain / Repository
     │
     ▼
PostgreSQL / Redis
     │
     ▼
Service
     │
     ▼
Response DTO
     │
     ▼
HTTP Response
```

## Key Technical Decisions

### Transactional invitation acceptance

Accepting a project invitation is performed inside a PostgreSQL transaction.

Adding the user to the project and updating the invitation status either **both succeed or both roll back**.

### Authorization at the database level

Access permissions are additionally enforced inside SQL queries.

Users can only interact with projects they belong to, and operations are restricted according to their project role.

Available roles:

```text
creator
admin
member
```

### Atomic refresh token rotation

Refresh sessions are stored in Redis.

Refresh token rotation is performed atomically, preventing the same refresh token from being successfully used by multiple concurrent requests.

If two requests attempt to rotate the same token simultaneously, only one succeeds.

### Graceful shutdown

When the application receives a shutdown signal, the HTTP server stops accepting new requests and gives active requests a limited amount of time to finish.

### Health and readiness checks

Two endpoints are available:

- `/health` — verifies that the HTTP application is running
- `/ready` — additionally verifies PostgreSQL and Redis connectivity

---

# Quick Start

## Requirements

Make sure the following tools are installed:

- Git
- Docker Desktop
- Docker Compose
- Go 1.27.2 (only when running checks outside Docker)

## 1. Clone the Repository

```bash
git clone https://github.com/glebH52-arch/petprojectjiro.git do-together
cd do-together
```

## 2. Configure Environment Variables

Create `.env` from the provided example:

### Linux / macOS

```bash
cp .env.example .env
```

### Windows PowerShell

```powershell
Copy-Item .env.example .env
```

At minimum, replace the following values:

```env
JWT_SECRET=replace_with_secret_at_least_32_characters
POSTGRES_PASSWORD=replace_with_postgres_password
REDIS_PASSWORD=replace_with_redis_password
EMAIL_VERIFICATION_SECRET=replace_with_another_secret_at_least_32_characters
```

> [!IMPORTANT]
> The `.env` file contains secrets and should never be committed to Git.

## 3. Start the Application

```bash
docker compose up -d --build
```

Docker Compose will automatically:

1. Start PostgreSQL
2. Start Redis
3. Wait for both services to become healthy
4. Apply PostgreSQL migrations
5. Build and start the API and email worker

By default `EMAIL_TRANSPORT=log` keeps the local demo self-contained: verification messages are written to the email-worker logs and no real email is sent. Read the current verification code with:

```bash
docker compose logs email-worker
```

To use Brevo instead, set `EMAIL_TRANSPORT=brevo`, `BREVO_API_KEY`, `BREVO_SENDER_EMAIL`, and optionally `BREVO_SENDER_NAME` in `.env`, then recreate the worker.

User avatars are stored in the Compose volume `avatar_data`, so recreating the API container does not remove them. Set `AVATAR_BASE_URL` to the public API origin when it differs from `http://localhost:8080`.

Check container status:

```bash
docker compose ps -a
```

PostgreSQL, Redis, and the API should report:

```text
healthy
```

The migration container should finish with:

```text
Exited (0)
```

## 4. Verify the Application

Check the HTTP server:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "ok"
}
```

Check application dependencies:

```bash
curl http://localhost:8080/ready
```

Expected response:

```json
{
  "status": "ready"
}
```

The API is available at `http://localhost:8080`.

## 5. Stop the Application

```bash
docker compose down
```

PostgreSQL and Redis data are preserved in Docker volumes.

---

# Configuration

The application is configured using environment variables.

| Variable | Description | Example |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection URL for local execution without Compose | `postgres://postgres:change_me@localhost:5432/do_together?sslmode=disable` |
| `JWT_SECRET` | Secret used for signing and verifying JWTs | `replace_with_secret_at_least_32_characters` |
| `ACCESS_TOKEN_TTL` | Access token lifetime | `15m` |
| `REDIS_ADDR` | Redis address for local execution | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | `change_me` |
| `REDIS_DB` | Redis logical database | `0` |
| `REFRESH_TOKEN_IDLE_TTL` | Refresh session inactivity timeout | `168h` |
| `REFRESH_TOKEN_ABSOLUTE_TTL` | Maximum refresh session lifetime | `1440h` |
| `POSTGRES_USER` | PostgreSQL user | `postgres` |
| `POSTGRES_PASSWORD` | PostgreSQL password | `change_me` |
| `POSTGRES_DB` | PostgreSQL database name | `do_together` |
| `API_PORT` | Host port for direct API access | `8080` |
| `EMAIL_VERIFICATION_SECRET` | Secret used to hash verification codes | `replace_with_another_secret_at_least_32_characters` |
| `EMAIL_VERIFICATION_CODE_TTL` | Verification code lifetime | `15m` |
| `EMAIL_TRANSPORT` | `log` for reproducible local email or `brevo` for delivery | `log` |
| `BREVO_API_KEY` | Brevo API key; required only for the `brevo` transport | empty |
| `BREVO_SENDER_EMAIL` | Verified Brevo sender; required only for the `brevo` transport | empty |
| `BREVO_SENDER_NAME` | Sender display name | `Do Together` |

When running with Docker Compose, PostgreSQL and Redis addresses are automatically configured to use:

```text
postgres:5432
redis:6379
```

---

# API

Protected endpoints require an access token:

```http
Authorization: Bearer <access_token>
```

## Authentication & Users

| Method | Endpoint | Description | Auth |
|---|---|---|:---:|
| `POST` | `/users` | Register a user | — |
| `POST` | `/auth/login` | Authenticate and receive tokens | — |
| `POST` | `/auth/refresh` | Rotate refresh token | — |
| `POST` | `/auth/logout` | Terminate refresh session | — |
| `GET` | `/users/me` | Get current user | JWT |

## Projects

| Method | Endpoint | Description | Auth |
|---|---|---|:---:|
| `POST` | `/projects` | Create a project | JWT |
| `GET` | `/projects` | List accessible projects | JWT |
| `GET` | `/projects/{id}` | Get project by ID | JWT |
| `PATCH` | `/projects/{id}` | Update project | JWT |
| `DELETE` | `/projects/{id}` | Archive project | JWT |

### Project List Parameters

`GET /projects` supports:

| Parameter | Description |
|---|---|
| `status` | `active` or `archived` |
| `limit` | Maximum number of returned projects |
| `offset` | Pagination offset |

## Project Members

| Method | Endpoint | Description | Auth |
|---|---|---|:---:|
| `GET` | `/projects/{id}/members` | List project members | JWT |
| `PATCH` | `/projects/{id}/members/{userID}` | Change member role | JWT |
| `DELETE` | `/projects/{id}/members/{userID}` | Remove member | JWT |

## Invitations

| Method | Endpoint | Description | Auth |
|---|---|---|:---:|
| `POST` | `/projects/{id}/invites` | Create invitation | JWT |
| `GET` | `/invites` | Get current user's invitations | JWT |
| `POST` | `/invites/{id}/accept` | Accept invitation | JWT |
| `POST` | `/invites/{id}/decline` | Decline invitation | JWT |

## Tasks

| Method | Endpoint | Description | Auth |
|---|---|---|:---:|
| `POST` | `/projects/{id}/tasks` | Create task | JWT |
| `GET` | `/projects/{id}/tasks` | List project tasks | JWT |
| `GET` | `/projects/{id}/tasks/{taskID}` | Get task by ID | JWT |
| `PATCH` | `/projects/{id}/tasks/{taskID}` | Update task or status | JWT |
| `DELETE` | `/projects/{id}/tasks/{taskID}` | Delete task | JWT |

## Health

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | HTTP application health |
| `GET` | `/ready` | PostgreSQL and Redis readiness |

---

# Usage Examples

All examples assume the API is running at:

```text
http://localhost:8080
```

## Register

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_user",
    "email": "demo@example.com",
    "password": "DemoPassword123!"
  }'
```

## Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@example.com",
    "password": "DemoPassword123!"
  }'
```

Example response:

```json
{
  "access_token": "<access_token>",
  "refresh_token": "<refresh_token>",
  "token_type": "Bearer",
  "expires_in": 900
}
```

## Create a Project

```bash
curl -X POST http://localhost:8080/projects \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Demo project",
    "goal": "Check the Do Together API"
  }'
```

## Get Projects

```bash
curl "http://localhost:8080/projects?status=active&limit=20&offset=0" \
  -H "Authorization: Bearer <access_token>"
```

## Refresh Tokens

```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh_token>"
  }'
```

After successful rotation, the old refresh token becomes invalid.

The client must store the **new refresh token** returned by the server and use it for the next refresh request.

## Logout

```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<current_refresh_token>"
  }'
```

A successful logout terminates the current refresh session and returns:

```text
204 No Content
```

---

# Testing & CI

The repository contains a GitHub Actions CI workflow.

The workflow runs on:

- pull requests targeting `main`;
- pushes to `main`.

It executes:

```bash
go mod download
go test ./...
go vet ./...
go build -o do-together ./cmd
```

## Automated Tests

Automated tests are currently located in:

```text
internal/domain/project_test.go
internal/repository/project_test.go
```

The remaining functionality has been manually tested through HTTP requests and Docker Compose, including:

- registration and authentication;
- refresh token rotation and logout;
- project creation and archiving;
- invitations;
- task management;
- role and status changes;
- concurrent refresh token rotation;
- `/health` and `/ready`;
- PostgreSQL and Redis persistence after container restarts.

---

# Current Limitations

Do Together is currently an **MVP**.

The following features are not yet implemented:

- Email verification
- Password recovery
- Email-based invitations — users are currently invited by user ID
- Rate limiting
- Frontend application
- Public deployment
- Comprehensive unit and integration test coverage

---

## Project Status

The project is under active development.

The current focus is on improving test coverage, authentication/session security, and expanding project management functionality.
