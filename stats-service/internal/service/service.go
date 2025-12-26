package service

import (
	"context"
	"time"

	"github.com/Domenick1991/students/stats-service/internal/repository"
)

type Service interface {
	GetDailyStats(ctx context.Context, date string) (int64, error)
	GetActiveRents(ctx context.Context) (int64, error)
	GetLocationStats(ctx context.Context, date string) (map[string]int64, error)
}

type service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetDailyStats(ctx context.Context, date string) (int64, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return s.repo.GetDailyStats(ctx, date)
}

func (s *service) GetActiveRents(ctx context.Context) (int64, error) {
	return s.repo.GetActiveRents(ctx)
}

func (s *service) GetLocationStats(ctx context.Context, date string) (map[string]int64, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	return s.repo.GetLocationStats(ctx, date)
}

