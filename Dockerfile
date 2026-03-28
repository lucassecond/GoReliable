# Stage 1 — compile the service with the full Go SDK and cache module downloads.
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy only module metadata first so dependency layers stay cached when app code changes.
COPY go.mod go.sum ./
RUN go mod download

# Copy application sources and produce a stripped release binary at /app.
COPY app/ ./app/
RUN go build -ldflags="-w -s" -o /app ./app

# Stage 2 — tiny image that runs only the compiled binary as a dedicated user.
FROM alpine:3.19

# Create a non-root account and drop privileges for runtime (no extra packages needed).
RUN adduser -D -u 1000 appuser

COPY --from=builder --chown=appuser:appuser /app /app

USER appuser

EXPOSE 8080

CMD ["/app"]
