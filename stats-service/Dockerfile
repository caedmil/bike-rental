FROM golang:latest AS builder

WORKDIR /app

# Git is already included in golang image
# Enable automatic toolchain selection
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Update go.mod and go.sum if needed
RUN go mod tidy || true
RUN go mod download

# Build the application
WORKDIR /app/stats-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/stats-service ./cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/stats-service .

EXPOSE 8081

CMD ["./stats-service"]

