package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"bike-rental/rent-service/internal/models"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRepository - мок для Repository интерфейса
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error) {
	args := m.Called(ctx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Bike), args.Error(1)
}

func (m *MockRepository) GetBikeByID(ctx context.Context, bikeID uuid.UUID) (*models.Bike, error) {
	args := m.Called(ctx, bikeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bike), args.Error(1)
}

func (m *MockRepository) StartRent(ctx context.Context, userID string, bikeID uuid.UUID) (*models.Rent, error) {
	args := m.Called(ctx, userID, bikeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rent), args.Error(1)
}

func (m *MockRepository) EndRent(ctx context.Context, rentID uuid.UUID, userID string) (*models.Rent, error) {
	args := m.Called(ctx, rentID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rent), args.Error(1)
}

func (m *MockRepository) GetRentByID(ctx context.Context, rentID uuid.UUID) (*models.Rent, error) {
	args := m.Called(ctx, rentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rent), args.Error(1)
}

func (m *MockRepository) UpdateBikeStatus(ctx context.Context, bikeID uuid.UUID, status string) error {
	args := m.Called(ctx, bikeID, status)
	return args.Error(0)
}

// MockKafkaWriter - мок для Kafka Writer интерфейса
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ServiceTestSuite - тестовый набор для Service
type ServiceTestSuite struct {
	suite.Suite
	mockRepo   *MockRepository
	mockWriter *MockKafkaWriter
	service    Service
	ctx        context.Context
}

// SetupTest - вызывается перед каждым тестом
func (suite *ServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockRepository)
	suite.mockWriter = new(MockKafkaWriter)
	suite.service = NewService(suite.mockRepo, suite.mockWriter, "test-topic")
	suite.ctx = context.Background()
}

// TearDownTest - вызывается после каждого теста
func (suite *ServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockWriter.AssertExpectations(suite.T())
}

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

// TestStartRent_InvalidBikeID - тест с невалидным bike_id
func (suite *ServiceTestSuite) TestStartRent_InvalidBikeID() {
	// Arrange
	userID := "user123"
	invalidBikeID := "invalid-uuid"

	// Act
	result, err := suite.service.StartRent(suite.ctx, userID, invalidBikeID)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "invalid bike_id")
}

// TestStartRent_RepositoryError - тест с ошибкой репозитория
func (suite *ServiceTestSuite) TestStartRent_RepositoryError() {
	// Arrange
	userID := "user123"
	bikeID := uuid.New()
	bikeIDStr := bikeID.String()
	
	expectedError := errors.New("bike is not available")
	suite.mockRepo.On("StartRent", suite.ctx, userID, bikeID).Return(nil, expectedError)

	// Act
	result, err := suite.service.StartRent(suite.ctx, userID, bikeIDStr)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(expectedError, err)
}

// TestStartRent_KafkaError - тест с ошибкой Kafka (не должно падать)
func (suite *ServiceTestSuite) TestStartRent_KafkaError() {
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
	suite.mockWriter.On("WriteMessages", suite.ctx, mock.Anything).Return(errors.New("kafka connection failed"))

	// Act
	result, err := suite.service.StartRent(suite.ctx, userID, bikeIDStr)

	// Assert - операция должна завершиться успешно несмотря на ошибку Kafka
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedRent.ID, result.ID)
}

// TestEndRent_Success - тест успешного завершения аренды
func (suite *ServiceTestSuite) TestEndRent_Success() {
	// Arrange
	userID := "user123"
	rentID := uuid.New()
	rentIDStr := rentID.String()
	bikeID := uuid.New()
	endTime := time.Now()
	
	expectedRent := &models.Rent{
		ID:        rentID,
		UserID:    userID,
		BikeID:    bikeID,
		StartTime: time.Now().Add(-1 * time.Hour),
		EndTime:   &endTime,
		Status:    "completed",
	}

	suite.mockRepo.On("EndRent", suite.ctx, rentID, userID).Return(expectedRent, nil)
	suite.mockWriter.On("WriteMessages", suite.ctx, mock.Anything).Return(nil)

	// Act
	result, err := suite.service.EndRent(suite.ctx, rentIDStr, userID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(rentID, result.ID)
	suite.Equal("completed", result.Status)
	suite.NotNil(result.EndTime)
}

// TestEndRent_InvalidRentID - тест с невалидным rent_id
func (suite *ServiceTestSuite) TestEndRent_InvalidRentID() {
	// Arrange
	userID := "user123"
	invalidRentID := "invalid-uuid"

	// Act
	result, err := suite.service.EndRent(suite.ctx, invalidRentID, userID)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "invalid rent_id")
}

// TestEndRent_RepositoryError - тест с ошибкой репозитория
func (suite *ServiceTestSuite) TestEndRent_RepositoryError() {
	// Arrange
	userID := "user123"
	rentID := uuid.New()
	rentIDStr := rentID.String()
	
	expectedError := errors.New("rent not found")
	suite.mockRepo.On("EndRent", suite.ctx, rentID, userID).Return(nil, expectedError)

	// Act
	result, err := suite.service.EndRent(suite.ctx, rentIDStr, userID)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(expectedError, err)
}

// TestGetAvailableBikes_Success - тест успешного получения доступных велосипедов
func (suite *ServiceTestSuite) TestGetAvailableBikes_Success() {
	// Arrange
	location := "Moscow"
	expectedBikes := []models.Bike{
		{
			ID:        uuid.New(),
			Name:      "Bike 1",
			Status:    "available",
			Location:  location,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "Bike 2",
			Status:    "available",
			Location:  location,
			CreatedAt: time.Now(),
		},
	}

	suite.mockRepo.On("GetAvailableBikes", suite.ctx, location).Return(expectedBikes, nil)

	// Act
	result, err := suite.service.GetAvailableBikes(suite.ctx, location)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 2)
	suite.Equal(expectedBikes[0].ID, result[0].ID)
	suite.Equal(expectedBikes[1].ID, result[1].ID)
}

// TestGetAvailableBikes_EmptyResult - тест с пустым результатом
func (suite *ServiceTestSuite) TestGetAvailableBikes_EmptyResult() {
	// Arrange
	location := "Unknown"
	expectedBikes := []models.Bike{}

	suite.mockRepo.On("GetAvailableBikes", suite.ctx, location).Return(expectedBikes, nil)

	// Act
	result, err := suite.service.GetAvailableBikes(suite.ctx, location)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Len(result, 0)
}

// TestGetAvailableBikes_RepositoryError - тест с ошибкой репозитория
func (suite *ServiceTestSuite) TestGetAvailableBikes_RepositoryError() {
	// Arrange
	location := "Moscow"
	expectedError := errors.New("database connection failed")

	suite.mockRepo.On("GetAvailableBikes", suite.ctx, location).Return(nil, expectedError)

	// Act
	result, err := suite.service.GetAvailableBikes(suite.ctx, location)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(expectedError, err)
}

// TestServiceTestSuite - запуск всего набора тестов
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

