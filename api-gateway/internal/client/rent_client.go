package client

import (
	"context"
	"fmt"

	"bike-rental/api-gateway/internal/models"
	"bike-rental/rent-service/proto/rent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RentClient interface {
	StartRent(ctx context.Context, userID, bikeID string) (*models.RentResponse, error)
	EndRent(ctx context.Context, rentID, userID string) (*models.RentResponse, error)
	GetAvailableBikes(ctx context.Context, location string) (*models.BikesList, error)
	Close() error
}

type rentClient struct {
	conn   *grpc.ClientConn
	client rent.RentServiceClient
}

func NewRentClient(address string) (RentClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rent service: %w", err)
	}

	return &rentClient{
		conn:   conn,
		client: rent.NewRentServiceClient(conn),
	}, nil
}

func (c *rentClient) StartRent(ctx context.Context, userID, bikeID string) (*models.RentResponse, error) {
	resp, err := c.client.StartRent(ctx, &rent.StartRentRequest{
		UserId: userID,
		BikeId: bikeID,
	})
	if err != nil {
		return nil, err
	}

	return &models.RentResponse{
		RentID:    resp.RentId,
		UserID:    resp.UserId,
		BikeID:    resp.BikeId,
		Status:    resp.Status,
		Message:   resp.Message,
		StartTime: resp.StartTime,
		EndTime:   resp.EndTime,
	}, nil
}

func (c *rentClient) EndRent(ctx context.Context, rentID, userID string) (*models.RentResponse, error) {
	resp, err := c.client.EndRent(ctx, &rent.EndRentRequest{
		RentId: rentID,
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	return &models.RentResponse{
		RentID:    resp.RentId,
		UserID:    resp.UserId,
		BikeID:    resp.BikeId,
		Status:    resp.Status,
		Message:   resp.Message,
		StartTime: resp.StartTime,
		EndTime:   resp.EndTime,
	}, nil
}

func (c *rentClient) GetAvailableBikes(ctx context.Context, location string) (*models.BikesList, error) {
	resp, err := c.client.GetAvailableBikes(ctx, &rent.AvailableBikesRequest{
		Location: location,
	})
	if err != nil {
		return nil, err
	}

	bikes := make([]models.Bike, 0, len(resp.Bikes))
	for _, b := range resp.Bikes {
		bikes = append(bikes, models.Bike{
			ID:       b.Id,
			Name:     b.Name,
			Status:   b.Status,
			Location: b.Location,
		})
	}

	return &models.BikesList{Bikes: bikes}, nil
}

func (c *rentClient) Close() error {
	return c.conn.Close()
}

