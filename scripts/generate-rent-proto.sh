#!/bin/bash

cd "$(dirname "$0")/.." || exit

# Генерация gRPC кода для rent service
protoc -I ./rent-service/proto \
  --go_out=./rent-service/proto/rent --go_opt=paths=source_relative \
  --go-grpc_out=./rent-service/proto/rent --go-grpc_opt=paths=source_relative \
  ./rent-service/proto/rent.proto

echo "Proto files generated successfully"

