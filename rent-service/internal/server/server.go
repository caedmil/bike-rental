package server

import (
	"context"

	"bike-rental/rent-service/internal/service"
	"bike-rental/rent-service/proto/rent"
)

type RentServer struct {
	rent.UnimplementedRentServiceServer
	service service.Service
}

func NewRentServer(svc service.Service) *RentServer {
	return &RentServer{
		service: svc,
	}
}

func (s *RentServer) StartRent(ctx context.Context, req *rent.StartRentRequest) (*rent.RentResponse, error) {
	rentModel, err := s.service.StartRent(ctx, req.UserId, req.BikeId)
	if err != nil {
		return &rent.RentResponse{
			Status:  "error",
			Message: err.Error(),
		}, nil
	}

	return &rent.RentResponse{
		RentId:    rentModel.ID.String(),
		UserId:    rentModel.UserID,
		BikeId:    rentModel.BikeID.String(),
		Status:    rentModel.Status,
		Message:   "Rent started successfully",
		StartTime: rentModel.StartTime.Unix(),
	}, nil
}

func (s *RentServer) EndRent(ctx context.Context, req *rent.EndRentRequest) (*rent.RentResponse, error) {
	rentModel, err := s.service.EndRent(ctx, req.RentId, req.UserId)
	if err != nil {
		return &rent.RentResponse{
			Status:  "error",
			Message: err.Error(),
		}, nil
	}

	var endTime int64
	if rentModel.EndTime != nil {
		endTime = rentModel.EndTime.Unix()
	}

	return &rent.RentResponse{
		RentId:    rentModel.ID.String(),
		UserId:    rentModel.UserID,
		BikeId:    rentModel.BikeID.String(),
		Status:    rentModel.Status,
		Message:   "Rent ended successfully",
		StartTime: rentModel.StartTime.Unix(),
		EndTime:   endTime,
	}, nil
}

func (s *RentServer) GetAvailableBikes(ctx context.Context, req *rent.AvailableBikesRequest) (*rent.BikesList, error) {
	bikes, err := s.service.GetAvailableBikes(ctx, req.Location)
	if err != nil {
		return nil, err
	}

	result := &rent.BikesList{
		Bikes: make([]*rent.Bike, 0, len(bikes)),
	}

	for _, b := range bikes {
		result.Bikes = append(result.Bikes, &rent.Bike{
			Id:       b.ID.String(),
			Name:     b.Name,
			Status:   b.Status,
			Location: b.Location,
		})
	}

	return result, nil
}

func (s *RentServer) GetRentStats(ctx context.Context, req *rent.StatsRequest) (*rent.StatsResponse, error) {
	// This will be handled by stats service
	// For now, return empty stats
	return &rent.StatsResponse{
		Date:        req.Date,
		TotalRents:  0,
		ActiveRents: 0,
		LocationStats: make(map[string]int64),
	}, nil
}

