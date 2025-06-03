# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments for cross-compilation
ARG TARGETOS
ARG TARGETARCH

# Build the application with cross-compilation
RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build -a -installsuffix cgo -ldflags '-w -s' -o gocoder .

# Final stage - Ubuntu with osmium tools
FROM ubuntu:22.04

# Install runtime dependencies including osmium tools
RUN apt-get update && apt-get install -y \
    ca-certificates \
    osmium-tool \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gocoder .

# Expose port (adjust as needed)
EXPOSE 8080

# Run the binary
CMD ["./gocoder", "server"]