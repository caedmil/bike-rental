package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"bike-rental/rent-service/internal/models"
	"bike-rental/rent-service/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ServiceTestSuite - тестовый набор для Service
type ServiceTestSuite struct {
	suite.Suite
	mockRepo   *mocks.Repository
	mockWriter *mocks.Writer
	service    Service
	ctx        context.Context
}

// SetupTest - вызывается перед каждым тестом
func (suite *ServiceTestSuite) SetupTest() {
	suite.mockRepo = mocks.NewRepository(suite.T())
	suite.mockWriter = mocks.NewWriter(suite.T())
	suite.service = NewService(suite.mockRepo, suite.mockWriter, "test-topic")
	suite.ctx = context.Background()
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

// TestAddBike_Success - тест успешного добавления велосипеда
func (suite *ServiceTestSuite) TestAddBike_Success() {
	// Arrange
	name := "Mountain Bike"
	location := "Park A"
	expectedBike := &models.Bike{
		ID:        uuid.New(),
		Name:      name,
		Status:    "available",
		Location:  location,
		CreatedAt: time.Now(),
	}

	suite.mockRepo.On("AddBike", suite.ctx, name, location).Return(expectedBike, nil)

	// Act
	result, err := suite.service.AddBike(suite.ctx, name, location)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(expectedBike.ID, result.ID)
	suite.Equal(name, result.Name)
	suite.Equal("available", result.Status)
	suite.Equal(location, result.Location)
}

// TestAddBike_EmptyName - тест с пустым именем велосипеда
func (suite *ServiceTestSuite) TestAddBike_EmptyName() {
	// Arrange
	name := ""
	location := "Park A"

	// Act
	result, err := suite.service.AddBike(suite.ctx, name, location)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "bike name is required")
}

// TestAddBike_EmptyLocation - тест с пустой локацией
func (suite *ServiceTestSuite) TestAddBike_EmptyLocation() {
	// Arrange
	name := "Mountain Bike"
	location := ""

	// Act
	result, err := suite.service.AddBike(suite.ctx, name, location)

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Contains(err.Error(), "bike location is required")
}

// TestDeleteBike_Success - тест успешного удаления велосипеда
func (suite *ServiceTestSuite) TestDeleteBike_Success() {
	// Arrange
	bikeID := uuid.New()
	bikeIDStr := bikeID.String()
	
	bike := &models.Bike{
		ID:        bikeID,
		Name:      "Test Bike",
		Status:    "available",
		Location:  "Park A",
		CreatedAt: time.Now(),
	}

	suite.mockRepo.On("HasActiveRent", suite.ctx, bikeID).Return(false, nil)
	suite.mockRepo.On("GetBikeByID", suite.ctx, bikeID).Return(bike, nil)
	suite.mockRepo.On("DeleteBike", suite.ctx, bikeID).Return(nil)

	// Act
	err := suite.service.DeleteBike(suite.ctx, bikeIDStr)

	// Assert
	suite.NoError(err)
}

// TestDeleteBike_InvalidUUID - тест с невалидным UUID
func (suite *ServiceTestSuite) TestDeleteBike_InvalidUUID() {
	// Arrange
	invalidBikeID := "invalid-uuid"

	// Act
	err := suite.service.DeleteBike(suite.ctx, invalidBikeID)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "invalid bike_id")
}

// TestDeleteBike_HasActiveRent - тест попытки удалить велосипед с активной арендой
func (suite *ServiceTestSuite) TestDeleteBike_HasActiveRent() {
	// Arrange
	bikeID := uuid.New()
	bikeIDStr := bikeID.String()

	suite.mockRepo.On("HasActiveRent", suite.ctx, bikeID).Return(true, nil)

	// Act
	err := suite.service.DeleteBike(suite.ctx, bikeIDStr)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "cannot delete bike: bike has active rent")
}

// TestDeleteBike_BikeNotFound - тест удаления несуществующего велосипеда
func (suite *ServiceTestSuite) TestDeleteBike_BikeNotFound() {
	// Arrange
	bikeID := uuid.New()
	bikeIDStr := bikeID.String()

	suite.mockRepo.On("HasActiveRent", suite.ctx, bikeID).Return(false, nil)
	suite.mockRepo.On("GetBikeByID", suite.ctx, bikeID).Return(nil, errors.New("bike not found"))

	// Act
	err := suite.service.DeleteBike(suite.ctx, bikeIDStr)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "bike not found")
}

// TestServiceTestSuite - запуск всего набора тестов
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}

