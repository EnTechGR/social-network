# ─────────────────────────────────────────────
# Stage 1 – Build
# ─────────────────────────────────────────────
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# CGO is required by go-sqlite3
ENV CGO_ENABLED=1 \
    GOOS=linux

# Install C toolchain (needed for CGO / SQLite)
RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc libc6-dev && \
    rm -rf /var/lib/apt/lists/*

# Cache module downloads as a separate layer
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Produce a statically-linkable binary
RUN go build -ldflags="-s -w" -o api ./cmd/main.go

# ─────────────────────────────────────────────
# Stage 2 – Runtime
# ─────────────────────────────────────────────
FROM debian:bookworm-slim

WORKDIR /app

# Runtime deps: CA certs for HTTPS calls, SQLite shared library, curl for healthcheck
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates libsqlite3-0 curl && \
    rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the builder stage
COPY --from=builder /app/api .

# Copy migration files – the app resolves them relative to the working dir
COPY --from=builder /app/pkg/db/migrations ./pkg/db/migrations

# Provide a default .env so the app doesn't fatal on startup;
# real secrets should be passed via docker-compose environment / secrets.
COPY --from=builder /app/.env.docker .env

# Persistent storage for the SQLite database and user uploads
VOLUME ["/app/database", "/app/uploads"]

EXPOSE 8080

CMD ["./api"]
