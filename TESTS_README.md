
**Файл:** `rent-service/internal/service/service_test.go`

**Покрытие:** 10 тестов для всех методов Service слоя


### Test Suite
```go
type ServiceTestSuite struct {
    suite.Suite
    mockRepo   *MockRepository
    mockWriter *MockKafkaWriter
    service    Service
    ctx        context.Context
}
```

### Моки

#### MockRepository
Мок для Repository интерфейса:
-  GetAvailableBikes
-  GetBikeByID
-  StartRent
-  EndRent
-  GetRentByID
-  UpdateBikeStatus

#### MockKafkaWriter
Мок для Kafka Writer интерфейса:
-  WriteMessages
-  Close

### Тесты (10 штук)

#### StartRent (4 теста)
1. `TestStartRent_Success` - успешный старт аренды
2.  `TestStartRent_InvalidBikeID` - невалидный bike_id
3.  `TestStartRent_RepositoryError` - ошибка репозитория
4.  `TestStartRent_KafkaError` - ошибка Kafka (не должна падать операцию)

#### EndRent (3 теста)
5. `TestEndRent_Success` - успешное завершение аренды
6.  `TestEndRent_InvalidRentID` - невалидный rent_id
7. `TestEndRent_RepositoryError` - ошибка репозитория

#### GetAvailableBikes (3 теста)
8.  `TestGetAvailableBikes_Success` - успешное получение списка
9.  `TestGetAvailableBikes_EmptyResult` - пустой результат
10.  `TestGetAvailableBikes_RepositoryError` - ошибка репозитория

### Установка зависимостей
```bash
go mod download
```

### Запуск всех тестов
```bash
go test ./rent-service/internal/service/... -v
```

### Запуск с покрытием
```bash
go test ./rent-service/internal/service/... -v -cover
```

### Генерация отчета о покрытии
```bash
go test ./rent-service/internal/service/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Запуск конкретного теста
```bash
go test ./rent-service/internal/service/... -v -run TestStartRent_Success
```

##  Ожидаемый результат

```
=== RUN   TestServiceTestSuite
=== RUN   TestServiceTestSuite/TestStartRent_Success
=== RUN   TestServiceTestSuite/TestStartRent_InvalidBikeID
=== RUN   TestServiceTestSuite/TestStartRent_RepositoryError
=== RUN   TestServiceTestSuite/TestStartRent_KafkaError
=== RUN   TestServiceTestSuite/TestEndRent_Success
=== RUN   TestServiceTestSuite/TestEndRent_InvalidRentID
=== RUN   TestServiceTestSuite/TestEndRent_RepositoryError
=== RUN   TestServiceTestSuite/TestGetAvailableBikes_Success
=== RUN   TestServiceTestSuite/TestGetAvailableBikes_EmptyResult
=== RUN   TestServiceTestSuite/TestGetAvailableBikes_RepositoryError
--- PASS: TestServiceTestSuite (0.01s)
    --- PASS: TestServiceTestSuite/TestStartRent_Success (0.00s)
    --- PASS: TestServiceTestSuite/TestStartRent_InvalidBikeID (0.00s)
    --- PASS: TestServiceTestSuite/TestStartRent_RepositoryError (0.00s)
    --- PASS: TestServiceTestSuite/TestStartRent_KafkaError (0.00s)
    --- PASS: TestServiceTestSuite/TestEndRent_Success (0.00s)
    --- PASS: TestServiceTestSuite/TestEndRent_InvalidRentID (0.00s)
    --- PASS: TestServiceTestSuite/TestEndRent_RepositoryError (0.00s)
    --- PASS: TestServiceTestSuite/TestGetAvailableBikes_Success (0.00s)
    --- PASS: TestServiceTestSuite/TestGetAvailableBikes_EmptyResult (0.00s)
    --- PASS: TestServiceTestSuite/TestGetAvailableBikes_RepositoryError (0.00s)
PASS
coverage: 85.7% of statements
ok      bike-rental/rent-service/internal/service       0.015s  coverage: 85.7% of statements
```

```go
// TestStartRent_Success - тест успешного старта аренды
func (suite *ServiceTestSuite) TestStartRent_Success() {
    // Arrange
    userID := "user123"
    bikeID := uuid.New()
    bikeIDStr := bikeID.String()
    
    expectedRent := &models.Rent{
        ID:        uuid.New(),
        UserID:    userID,
        BikeID:    bikeID,
        StartTime: time.Now(),
        Status:    "active",
    }

    suite.mockRepo.On("StartRent", suite.ctx, userID, bikeID).Return(expectedRent, nil)
    suite.mockWriter.On("WriteMessages", suite.ctx, mock.Anything).Return(nil)

    // Act
    result, err := suite.service.StartRent(suite.ctx, userID, bikeIDStr)

    // Assert
    suite.NoError(err)
    suite.NotNil(result)
    suite.Equal(expectedRent.ID, result.ID)
    suite.Equal(userID, result.UserID)
    suite.Equal(bikeID, result.BikeID)
    suite.Equal("active", result.Status)
}
```


