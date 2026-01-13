FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o avg-weights-server ./cmd/server/main.go



FROM python:3.11-slim AS runtime

WORKDIR /app

# ---- system deps for torch ----
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

# ---- install torch (CPU) ----
RUN pip install --no-cache-dir torch

# ---- copy python code ----
COPY internal/cron/aggregator/fedavg.py ./fedavg.py

# ---- copy Go binary ----
COPY --from=builder /app/avg-weights-server ./avg-weights-server

EXPOSE 8081

CMD ["./avg-weights-server"]
