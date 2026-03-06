# ─────────────────────────────────────────────
# Stage 1 – ALL dependencies (dev + prod)
#   gsap, TypeScript types, and other devDeps
#   are required by the compiler at build time.
# ─────────────────────────────────────────────
FROM node:20-bookworm-slim AS deps

WORKDIR /app

COPY package.json package-lock.json* ./
# Install everything – devDependencies included
RUN npm ci

# ─────────────────────────────────────────────
# Stage 2 – Build
# ─────────────────────────────────────────────
FROM node:20-bookworm-slim AS builder

WORKDIR /app

# Suppress the interactive telemetry notice during build
ENV NEXT_TELEMETRY_DISABLED=1

# Copy all node_modules (dev + prod) from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source
COPY . .

# NEXT_PUBLIC_* vars are inlined at build time.
# The browser talks to the backend via the HOST-exposed port,
# so we default to localhost:8080. Override with --build-arg if needed.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}

RUN npm run build

# ─────────────────────────────────────────────
# Stage 3 – Runtime
# ─────────────────────────────────────────────
FROM node:20-bookworm-slim AS runner

WORKDIR /app

ENV NODE_ENV=production

# Only copy what the Next.js server actually needs at runtime
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 8081

# Next.js standalone server listens on PORT env var (default 3000).
# We override it to 8081 to match the expected frontend port.
ENV PORT=8081
ENV HOSTNAME=0.0.0.0

CMD ["node", "server.js"]
