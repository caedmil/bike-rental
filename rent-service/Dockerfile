FROM golang:latest AS builder

WORKDIR /app

# Git is already included in golang image
# Enable automatic toolchain selection
ENV GOTOOLCHAIN=auto

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install protoc and generate proto files
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*

# Install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Copy source code
COPY . .

# Generate proto files
WORKDIR /app
ENV PATH=$PATH:/root/go/bin
RUN protoc -I ./rent-service/proto -I ./api \
    --go_out=./rent-service/proto/rent --go_opt=paths=source_relative \
    --go-grpc_out=./rent-service/proto/rent --go-grpc_opt=paths=source_relative \
    ./rent-service/proto/rent.proto

# Update go.mod and go.sum if needed
RUN go mod tidy || true

# Build the application
WORKDIR /app/rent-service
RUN CGO_ENABLED=0 GOOS=linux go build -o rent-service ./cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

COPY --from=builder /app/rent-service/rent-service .

EXPOSE 50051

CMD ["./rent-service"]

