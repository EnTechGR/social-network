# Social Network

Full-stack social network application with a Go backend API and a Next.js frontend.

## Overview

This repository contains:

- `backend/API`: Go API server, SQLite storage, migrations, WebSocket support, Swagger docs
- `frontend`: Next.js 16 application (React 19, TypeScript)
- `docker-compose.yml`: One-command local orchestration for backend + frontend

Core features include authentication, profiles, follows, posts/comments/reactions, groups, events, notifications, and chat.

## Tech Stack

- Backend: Go `1.24`, `net/http`, SQLite, `golang-migrate`, Gorilla WebSocket, Swagger
- Frontend: Next.js `16`, React `19`, TypeScript, Tailwind CSS `v4`
- Infra: Docker Compose, multi-stage Docker builds

## Repository Structure

```text
.
├── backend/
│   └── API/
│       ├── cmd/main.go
│       ├── pkg/
│       │   ├── db/migrations/
│       │   ├── handlers/
│       │   ├── middleware/
│       │   ├── models/
│       │   ├── repository/
│       │   ├── routes/
│       │   ├── utils/
│       │   └── websocket/
│       └── docs/
├── frontend/
│   ├── app/
│   ├── components/
│   ├── lib/
│   └── types/
├── docker-compose.yml
└── docker-guide.md
```

## Quick Start (Docker)

Prerequisites:

- Docker + Docker Compose plugin

Run from repository root:

```bash
docker compose up --build
```

Services:

- Frontend: `http://localhost:8081`
- Backend API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`

Useful commands:

```bash
# Start in background
docker compose up -d --build

# Follow logs
docker compose logs -f

# Stop services
docker compose down

# Stop and remove persisted data volumes
docker compose down -v
```

## Local Development (Without Docker)

### 1. Backend

Prerequisites:

- Go `1.24+`
- C toolchain for `go-sqlite3` (for example `gcc`)

Run:

```bash
cd backend/API
go mod download
go run ./cmd/main.go
```

Backend defaults to `http://localhost:8080`.

Environment:

- Backend loads variables from `backend/API/.env`
- For Docker image defaults, `backend/API/.env.docker` is copied into the container at build time
- OAuth-related variables (`GOOGLE_*`, `GITHUB_*`) should be set for social login flows

### 2. Frontend

Prerequisites:

- Node.js `20+`
- npm

Run:

```bash
cd frontend
npm ci
npm run dev
```

Frontend runs on `http://localhost:8081`.

Environment:

- Browser API URL is read from `NEXT_PUBLIC_API_URL` (defaults to `http://localhost:8080`)

## API Notes

- API base path: `/api/v1`
- Session verify endpoint: `GET /api/v1/verify`
- WebSocket endpoint: `/ws`
- Swagger docs: `/swagger/index.html`

## Testing

Backend tests:

```bash
cd backend/API
go test ./...
```

Frontend lint:

```bash
cd frontend
npm run lint
```

## Additional Docs

- Docker setup details: `docker-guide.md`
- Frontend-specific notes: `frontend/README.md`

## License

See `LICENSE`.
