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

# Build the application
WORKDIR /app/rent-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/rent-service ./cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/rent-service .

EXPOSE 50051

CMD ["./rent-service"]

