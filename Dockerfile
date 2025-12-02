FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and other dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations for cross-compilation
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o helmchecker ./cmd/helmchecker

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates git openssh-client

# Create a non-root user with specific UID/GID
RUN adduser -D -s /bin/sh -u 1000 helmchecker

WORKDIR /app

# Copy the binary from builder stage with execute permissions
COPY --from=builder --chmod=755 /app/helmchecker .

# Create directory for SSH keys that can be accessed by the non-root user
RUN mkdir -p /home/helmchecker/.ssh && chmod 700 /home/helmchecker/.ssh && chown -R helmchecker:helmchecker /home/helmchecker

# Switch to non-root user
USER helmchecker

ENTRYPOINT ["./helmchecker"]