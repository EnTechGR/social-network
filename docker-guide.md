# Docker Setup Guide

## Overview

The project is split into two containers that run on the same Docker bridge network (`social-network`):

| Container | Technology | Internal hostname | Host port |
|---|---|---|---|
| `social-network-backend` | Go + SQLite | `backend` | `8080` |
| `social-network-frontend` | Next.js 16 | `frontend` | `8081` |

Because both containers share the same Docker network, the Next.js server-side code can reach the backend using `http://backend:8080`, while the user's **browser** reaches it through the host-exposed port `http://localhost:8080`.

---

## File Placement

Place each file at the path shown relative to the **project root**:

```
project-root/
├── docker-compose.yml                ← orchestrates both services
├── backend/
│   └── API/
│       ├── Dockerfile                ← Go multi-stage build
│       └── .env.docker               ← default env for Docker image
└── frontend/
    ├── Dockerfile                    ← Next.js multi-stage build
    └── next.config.ts                ← must enable "output: standalone"
```

---

## Required: Enable Next.js Standalone Output

The frontend `Dockerfile` relies on Next.js **standalone** output mode.  
Add `output: 'standalone'` to `frontend/next.config.ts`:

```ts
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',      // ← add this line
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8080',
        pathname: '/**',
      },
    ],
    unoptimized: true,
  },
};

export default nextConfig;
```

---

## Quick Start

### 1. Prepare the backend `.env.docker`

Create `backend/API/.env.docker` (already tracked in version control):

```
FRONTEND_ORIGIN=http://localhost:8081
COOKIE_SECURE=false
APP_ENV=development
```

> **Note:** The real `backend/API/.env` is in `.gitignore` and will not be present in CI or fresh clones. The Dockerfile copies `.env.docker` as `.env` so the Go binary can start without crashing.

### 2. Build and start everything

From the **project root**:

```bash
docker compose up --build
```

On subsequent starts (no code changes):

```bash
docker compose up
```

### 3. Verify

| What | URL |
|---|---|
| Frontend | http://localhost:8081 |
| Backend API root | http://localhost:8080/api/v1 |
| Backend health | http://localhost:8080/api/v1/health |

---

## Common Commands

```bash
# Start in the background
docker compose up -d --build

# Watch logs from both services
docker compose logs -f

# Watch only the backend
docker compose logs -f backend

# Rebuild a single service after code changes
docker compose up --build backend

# Stop everything (keeps volumes)
docker compose down

# Stop and delete all data (SQLite DB + uploads)
docker compose down -v

# Open a shell inside the backend container
docker compose exec backend /bin/bash

# Open a shell inside the frontend container
docker compose exec frontend /bin/sh
```

---

## Environment Variables Reference

### Backend (`backend/API/.env.docker` or `docker-compose.yml` → `environment:`)

| Variable | Default | Description |
|---|---|---|
| `FRONTEND_ORIGIN` | `http://localhost:8081` | Allowed CORS origin for the browser frontend |
| `COOKIE_SECURE` | `false` | Set to `true` only when serving over HTTPS |
| `APP_ENV` | `development` | Set to `production` to tighten cookie/security flags |

### Frontend (build args in `docker-compose.yml` → `args:`)

| Variable | Default | Description |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend URL as seen by the **browser** (inlined at build time) |
| `INTERNAL_API_URL` | `http://backend:8080` | Backend URL for server-side Next.js requests (runtime env) |

---

## Networking Explained

```
┌────────────────────────────────────────────────────────┐
│                Docker bridge: social-network            │
│                                                        │
│  ┌──────────────────┐        ┌──────────────────────┐  │
│  │    backend       │◄───────│      frontend        │  │
│  │  :8080 (Go API)  │        │  :8081 (Next.js)     │  │
│  └────────┬─────────┘        └──────────┬───────────┘  │
│           │                             │               │
└───────────┼─────────────────────────────┼───────────────┘
            │ host port 8080              │ host port 8081
            ▼                             ▼
       Browser calls                 Browser loads
       API directly                  the frontend
```

- **Browser → Backend**: `http://localhost:8080` (host port, CORS-allowed)
- **Frontend SSR → Backend**: `http://backend:8080` (internal Docker DNS)
- **CORS**: The backend allows `http://localhost:8081` via `FRONTEND_ORIGIN`

---

## Persistent Data (Volumes)

| Volume | Mount path | Contents |
|---|---|---|
| `backend-db` | `/app/database` | SQLite database file |
| `backend-uploads` | `/app/uploads` | User-uploaded images |

Volumes survive `docker compose down` but are removed with `docker compose down -v`.  
Back up the `backend-db` volume before destructive operations:

```bash
docker run --rm \
  -v social-network_backend-db:/data \
  -v $(pwd):/backup \
  busybox tar czf /backup/db-backup.tar.gz /data
```

---

## Production Checklist

Before going to production, update the following:

1. **`NEXT_PUBLIC_API_URL`** → replace `localhost` with your public API domain/IP.
2. **`FRONTEND_ORIGIN`** → replace `localhost:8081` with your public frontend domain.
3. **`COOKIE_SECURE`** → set to `true` (requires HTTPS).
4. **`APP_ENV`** → set to `production`.
5. Place real secrets (JWT secret, OAuth keys, etc.) in `backend/API/.env`, never in the Dockerfile or version control.
6. Add a reverse proxy (nginx / Caddy) in front of both services to handle TLS termination.

---

## Troubleshooting

### Frontend shows a blank page / API calls fail

The `NEXT_PUBLIC_API_URL` build arg is inlined into the JS bundle at **build time**.  
If you change it, you must rebuild the frontend image:

```bash
docker compose up --build frontend
```

### Backend exits immediately

Check logs:

```bash
docker compose logs backend
```

The most common cause is a missing or malformed `.env.docker`. Ensure it exists at `backend/API/.env.docker` with at least:

```
FRONTEND_ORIGIN=http://localhost:8081
```

### SQLite "unable to open database"

The database path is relative to the working directory inside the container (`/app`).  
Ensure the `backend-db` volume is mounted and the container user has write permission.

### Port already in use

```bash
# Find the process using port 8080 or 8081
lsof -i :8080
lsof -i :8081
```

Change the left-hand side of the port mapping in `docker-compose.yml` if needed, e.g. `"8082:8080"`.