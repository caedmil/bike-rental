# Тестирование Rent Service

## Запуск тестов

### Все тесты
```bash
go test ./rent-service/internal/service/... -v
```

### С покрытием
```bash
go test ./rent-service/internal/service/... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Конкретный тест
```bash
go test ./rent-service/internal/service/... -v -run TestStartRent_Success
```

## Структура тестов

Тесты используют **testify/suite** и **testify/mock** для:
- Изоляции тестов
- Моков Repository и Kafka Writer
- Проверки всех сценариев

### Покрытие тестами

**rent-service/internal/service/service_test.go** покрывает:

1.  `StartRent`:
   - Успешный старт аренды
   - Невалидный bike_id
   - Ошибка репозитория
   - Ошибка Kafka (не должна падать операцию)

2.  `EndRent`:
   - Успешное завершение аренды
   - Невалидный rent_id
   - Ошибка репозитория

3.  `GetAvailableBikes`:
   - Успешное получение списка
   - Пустой результат
   - Ошибка репозитория

## Архитектура (слои и интерфейсы)

### Слой Service
- **Интерфейс**: `Service` (service.go:16-20)
- **Реализация**: `service` struct
- **Зависимости**: Repository (интерфейс), Kafka Writer (интерфейс)

### Слой Repository
- **Интерфейс**: `Repository` (repository.go:14-21)
- **Реализация**: `repository` struct
- **Зависимость**: PostgreSQL pool

### Слой Kafka
- **Интерфейс**: `Writer` (kafka/writer.go:9-12)
- **Реализация**: `KafkaWriter` struct
- **Зависимость**: segmentio/kafka-go Writer

### Слой Server (gRPC)
- **Реализация**: `RentServer` struct (server/server.go)
- **Зависимость**: Service (интерфейс)

## Моки в тестах

### MockRepository
Мок для Repository интерфейса с методами:
- `GetAvailableBikes`
- `StartRent`
- `EndRent`
- `GetBikeByID`
- `GetRentByID`
- `UpdateBikeStatus`

### MockKafkaWriter
Мок для Kafka Writer интерфейса с методами:
- `WriteMessages`
- `Close`

## Пример теста

```go
func (suite *ServiceTestSuite) TestStartRent_Success() {
    // Arrange
    userID := "user123"
    bikeID := uuid.New()
    
    expectedRent := &models.Rent{
        ID:        uuid.New(),
        UserID:    userID,
        BikeID:    bikeID,
        Status:    "active",
    }

    suite.mockRepo.On("StartRent", suite.ctx, userID, bikeID).Return(expectedRent, nil)
    suite.mockWriter.On("WriteMessages", suite.ctx, mock.Anything).Return(nil)

    // Act
    result, err := suite.service.StartRent(suite.ctx, userID, bikeID.String())

    // Assert
    suite.NoError(err)
    suite.NotNil(result)
    suite.Equal(expectedRent.ID, result.ID)
}
```



