# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make python3 py3-pip curl

WORKDIR /build

# Set GOTOOLCHAIN to auto to allow downloading required Go version
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM python:3.11-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python packages for FedAvg
RUN pip install --no-cache-dir torch==2.1.2+cpu --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir numpy

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server /app/server

# Copy Python script for FedAvg
COPY internal/cron/aggregator/fedavg.py /app/internal/cron/aggregator/fedavg.py

# Expose gRPC port
EXPOSE 8081

# Run the server
CMD ["/app/server"]

