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

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/helmchecker .

# Create directory for SSH keys
RUN mkdir -p ~/.ssh && chmod 700 ~/.ssh

ENTRYPOINT ["./helmchecker"]