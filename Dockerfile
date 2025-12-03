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

# Final stage - using distroless for minimal attack surface
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the binary from builder stage with execute permissions
COPY --from=builder --chmod=755 /app/helmchecker .

# distroless/static-debian12:nonroot already runs as nonroot user (UID 65532)
# and includes ca-certificates

ENTRYPOINT ["./helmchecker"]