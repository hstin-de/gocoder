FROM golang:1.23-bullseye AS builder

# Install build dependencies
RUN apt-get update && \
    apt-get install -y pkg-config zlib1g-dev && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o geocoder .

FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && \
    apt-get install -y osmium-tool sqlite3 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary
COPY --from=builder /app/geocoder .

# Create directories for data
RUN mkdir -p /data/maps /data/database

# Expose the server port
EXPOSE 3000

CMD ["./geocoder", "server"]