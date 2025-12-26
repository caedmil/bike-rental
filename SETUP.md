# Инструкция по настройке проекта

## Предварительные требования

1. Go 1.21 или выше
2. Docker и Docker Compose
3. Protoc (для генерации proto файлов) - опционально

## Установка зависимостей

```bash
go mod download
go mod tidy
```

## Генерация Proto файлов (опционально)

Если вы хотите перегенерировать proto файлы:

```bash
# Убедитесь, что protoc установлен
# Затем выполните:
./scripts/generate-rent-proto.sh
```

Или вручную:

```bash
protoc -I ./rent-service/proto \
  -I ./api/google/api \
  --go_out=./rent-service/proto --go_opt=paths=source_relative \
  --go-grpc_out=./rent-service/proto --go-grpc_opt=paths=source_relative \
  ./rent-service/proto/rent.proto
```

## Запуск проекта

### 1. Запуск инфраструктуры

```bash
docker-compose up -d postgres redis kafka
```

Подождите несколько секунд, пока все сервисы запустятся.

### 2. Запуск всех сервисов

```bash
docker-compose up -d
```

### 3. Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Получить доступные велосипеды
curl http://localhost:8080/api/v1/bikes/available
```

## Swagger документация

После запуска API Gateway, Swagger UI доступен по адресу:
- http://localhost:8080/docs/

## Структура проекта

Проект состоит из трех микросервисов:

1. **api-gateway** - HTTP API Gateway с Swagger
2. **rent-service** - gRPC сервис для управления арендой
3. **stats-service** - HTTP сервис для статистики

## Конфигурация

Все настройки находятся в `config.yaml`. Для Docker окружения используются имена сервисов из docker-compose.

## Локальная разработка

Для локальной разработки без Docker:

1. Обновите `config.yaml` для использования `localhost` вместо имен сервисов
2. Убедитесь, что PostgreSQL, Redis и Kafka запущены локально
3. Запустите сервисы в отдельных терминалах:

```bash
# Rent Service
cd rent-service && go run cmd/main.go

# Stats Service  
cd stats-service && go run cmd/main.go

# API Gateway
cd api-gateway && go run cmd/main.go
```

