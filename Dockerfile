# Build stage
FROM golang:1.24-bookworm AS builder

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Build the main binary
RUN CGO_ENABLED=0 GOOS=linux go build -o argocd-konn-jsonnet-plugin

# Build the go-jsonnet binary
RUN go install github.com/google/go-jsonnet/cmd/jsonnet@latest

# Final stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy binaries from builder stage
COPY --from=builder /app/argocd-konn-jsonnet-plugin /usr/local/bin/
COPY --from=builder /go/bin/jsonnet /usr/local/bin/

# Set the binary as entrypoint
ENTRYPOINT ["argocd-konn-jsonnet-plugin"]
