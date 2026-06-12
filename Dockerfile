# ============================================================================
# QuickLens - Multi-stage Dockerfile
# ============================================================================
# Stage 1: node:20-slim   → Build SvelteKit frontend
# Stage 2: golang:1.24    → Build static Go binary
# Stage 3: alpine:3.19    → Minimal runtime (~25 MB)
# ============================================================================

# Stage 1: Build Frontend
FROM node:20-slim AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --prefer-offline
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.24-alpine AS backend-builder
RUN apk add --no-cache git gcc musl-dev
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-builder /app/backend/frontend/build ./internal/static/build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o quicklens-backend ./cmd/server

# Stage 3: Runtime
FROM alpine:3.19 AS runtime
RUN apk add --no-cache curl ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/backend/quicklens-backend ./quicklens-backend
RUN mkdir -p /app/data
EXPOSE 8000
ENV PORT=8000 SQLITE_DB_PATH=/app/data/quicklens.db
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
  CMD curl -f http://localhost:8000/health || exit 1
CMD ["./quicklens-backend"]
