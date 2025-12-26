#!/bin/bash

cd "$(dirname "$0")/.." || exit

# Генерация gRPC кода для rent service
protoc -I ./rent-service/proto \
  -I ./api/google/api \
  --go_out=./rent-service/proto --go_opt=paths=source_relative \
  --go-grpc_out=./rent-service/proto --go-grpc_opt=paths=source_relative \
  ./rent-service/proto/rent.proto

# Генерация gRPC-Gateway
protoc -I ./rent-service/proto \
  -I ./api/google/api \
  --grpc-gateway_out=./rent-service/proto \
  --grpc-gateway_opt paths=source_relative \
  --grpc-gateway_opt logtostderr=true \
  ./rent-service/proto/rent.proto

# Генерация OpenAPI
protoc -I ./rent-service/proto \
  -I ./api/google/api \
  --openapiv2_out=./rent-service/proto/swagger \
  --openapiv2_opt logtostderr=true \
  ./rent-service/proto/rent.proto

echo "Proto files generated successfully"

