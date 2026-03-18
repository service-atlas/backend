# syntax=docker/dockerfile:1

########################
# Build stage
########################
FROM golang:1.26.1 AS builder
LABEL authors="joshp"

# Ensure modules are on and build is reproducible
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the rest of the source
COPY . .

# Build the service binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o /out/service-atlas ./cmd/service-atlas

########################
# Runtime stage (scratch)
########################
FROM scratch AS runtime

# Copy CA certificates for TLS (e.g., Neo4j over TLS or outbound HTTPS)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the statically linked binary
COPY --from=builder /out/service-atlas /service-atlas

# Non-root (nobody) user ID commonly available; scratch has no /etc/passwd, but UID works
USER 65532:65532

# Expose default HTTP port (adjust if you configure a different one)
EXPOSE 8080

# Run it
ENTRYPOINT ["/service-atlas"]

