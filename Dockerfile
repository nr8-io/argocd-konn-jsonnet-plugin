# Build stage
FROM golang:1.24-alpine AS builder

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
FROM alpine:latest

RUN apk add --no-cache git

# Copy binaries from builder stage
COPY --from=builder /app/argocd-konn-jsonnet-plugin /usr/local/bin/
COPY --from=builder /go/bin/jsonnet /usr/local/bin/

# Define group and user IDs for consistency
ENV GROUP_ID=999
ENV USER_ID=999

# Group and user needed for argocd-cmp-server
RUN grep 999 /etc/group | awk -F: '{print $1}' | xargs delgroup || true && \
    addgroup -g 999 argo && \
    adduser -D -u 999 -G argo argo

# set the default user
USER argo

# argocd-cmp-server as entrypoint
ENTRYPOINT ["/var/run/argocd/argocd-cmp-servers"]
