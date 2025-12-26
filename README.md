# Bike Rental System

Микросервисная система аренды велосипедов на Go с использованием PostgreSQL, Redis и Kafka.

## Архитектура

```
User → API Gateway (HTTP) → Rent Service (gRPC) → PostgreSQL
                             ↓
                           Kafka → Stats Service → Redis
```

## Компоненты

1. **API Gateway** (порт 8080) - HTTP API с Swagger документацией
2. **Rent Service** (gRPC порт 50051) - Сервис управления арендой
3. **Stats Service** (HTTP порт 8081) - Сервис статистики
4. **Kafka UI** (порт 8082) - Веб-интерфейс для просмотра топиков и сообщений Kafka
5. **Redis Commander** (порт 8083) - Веб-интерфейс для просмотра данных в Redis

## Технологии

- Go 1.21+
- PostgreSQL 15
- Redis 7
- Apache Kafka
- gRPC
- Docker & Docker Compose

## Быстрый старт

### 1. Запуск инфраструктуры

```bash
docker-compose up -d postgres redis kafka
```

### 2. Инициализация базы данных

База данных инициализируется автоматически при первом запуске PostgreSQL через `scripts/init-db.sql`.

### 3. Запуск всех сервисов

```bash
docker-compose up -d
```

### 4. Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Получить доступные велосипеды
curl http://localhost:8080/api/v1/bikes/available

# Начать аренду
curl -X POST http://localhost:8080/api/v1/rent/start \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user1", "bike_id": "bike-id-from-previous-request"}'

# Завершить аренду
curl -X POST http://localhost:8080/api/v1/rent/end \
  -H "Content-Type: application/json" \
  -d '{"rent_id": "rent-id-from-start", "user_id": "user1"}'

# Получить статистику
curl http://localhost:8080/api/v1/stats/active
curl http://localhost:8080/api/v1/stats/daily/2024-01-01
```

## Swagger документация

После запуска API Gateway, Swagger UI доступен по адресу:
- http://localhost:8080/docs/

## API Endpoints

### API Gateway (HTTP :8080)

- `POST /api/v1/rent/start` - Начать аренду
- `POST /api/v1/rent/end` - Завершить аренду
- `GET /api/v1/bikes/available` - Получить доступные велосипеды
- `GET /api/v1/stats/daily/{date}` - Статистика за день
- `GET /api/v1/stats/active` - Количество активных аренд
- `GET /health` - Health check
- `GET /docs/` - Swagger UI

### Stats Service (HTTP :8081)

- `GET /internal/stats/daily` - Статистика за день
- `GET /internal/stats/active` - Активные аренды
- `POST /admin/refresh-stats` - Обновить статистику

### Rent Service (gRPC :50051)

- `StartRent` - Начать аренду
- `EndRent` - Завершить аренду
- `GetAvailableBikes` - Получить доступные велосипеды
- `GetRentStats` - Получить статистику аренды

## Структура проекта

```
.
├── api-gateway/          # API Gateway сервис
│   ├── cmd/
│   ├── internal/
│   │   ├── handlers/    # HTTP handlers
│   │   ├── client/      # gRPC и HTTP клиенты
│   │   └── models/      # Модели данных
│   └── Dockerfile
├── rent-service/         # Rent Service
│   ├── cmd/
│   ├── internal/
│   │   ├── service/     # Бизнес-логика
│   │   ├── repository/  # Работа с БД
│   │   ├── server/      # gRPC server
│   │   └── models/      # Модели данных
│   ├── proto/           # Proto файлы
│   └── Dockerfile
├── stats-service/        # Stats Service
│   ├── cmd/
│   ├── internal/
│   │   ├── consumer/    # Kafka consumer
│   │   ├── repository/  # Redis repository
│   │   ├── service/     # Бизнес-логика
│   │   └── handlers/    # HTTP handlers
│   └── Dockerfile
├── config/              # Конфигурация
├── scripts/             # Скрипты инициализации
├── docker-compose.yaml  # Docker Compose конфигурация
└── config.yaml          # Конфигурационный файл
```

## Конфигурация

Все настройки находятся в файле `config.yaml`:

```yaml
database:
  postgres:
    host: "postgres"
    port: 5432
    user: "user"
    password: "pass"
    dbname: "bikerent"
  
  redis:
    address: "redis:6379"
    password: ""
    db: 0

kafka:
  brokers: ["kafka:9094"]
  topics:
    rent_events: "bike-rent-events"
    status_events: "bike-status-events"

services:
  rent_service: "rent-service:50051"
```

## Разработка

### Генерация proto файлов

```bash
protoc --go_out=. --go-grpc_out=. rent-service/proto/rent.proto
```

### Локальный запуск (без Docker)

1. Убедитесь, что PostgreSQL, Redis и Kafka запущены
2. Обновите `config.yaml` для локальных подключений
3. Запустите сервисы:

```bash
# Rent Service
cd rent-service && go run cmd/main.go

# Stats Service
cd stats-service && go run cmd/main.go

# API Gateway
cd api-gateway && go run cmd/main.go
```

## Мониторинг

### Kafka UI
Доступен по адресу: **http://localhost:8082**

Позволяет:
- Просматривать все топики Kafka
- Видеть сообщения в реальном времени
- Мониторить consumer groups
- Просматривать метаданные брокеров

### Redis Commander
Доступен по адресу: **http://localhost:8083**

Учетные данные:
- Username: `admin`
- Password: `admin`

Позволяет:
- Просматривать все ключи в Redis
- Видеть значения в реальном времени
- Мониторить статистику Redis
- Выполнять команды Redis

### Health Checks
Доступны на всех сервисах:
- API Gateway: http://localhost:8080/health
- Stats Service: http://localhost:8081/health
- Rent Service: gRPC на порту 50051

## 🧪 Тестирование

### Запуск тестов

```bash
# Все тесты
go test ./rent-service/internal/service/... -v

# С покрытием
go test ./rent-service/internal/service/... -v -cover

# Генерация HTML отчета
go test ./rent-service/internal/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Структура тестов

**Файл:** `rent-service/internal/service/service_test.go`

**Покрытие:** 10 тестов для Service слоя:
- StartRent (4 теста: success, invalid ID, repo error, kafka error)
- EndRent (3 теста: success, invalid ID, repo error)
- GetAvailableBikes (3 теста: success, empty, repo error)

**Технологии:**
- testify/suite - Test Suites
- testify/mock - Моки для Repository и Kafka Writer
- testify/assert - Assertions

 **Подробная документация:** [TESTS_README.md](TESTS_README.md)

MIT

