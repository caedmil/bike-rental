package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Repository interface {
	IncrementDailyRent(ctx context.Context, date string) error
	IncrementActiveRents(ctx context.Context) error
	DecrementActiveRents(ctx context.Context) error
	IncrementLocationRent(ctx context.Context, date string, location string) error
	GetDailyStats(ctx context.Context, date string) (int64, error)
	GetActiveRents(ctx context.Context) (int64, error)
	GetLocationStats(ctx context.Context, date string) (map[string]int64, error)
}

type repository struct {
	client *redis.Client
}

func NewRepository(client *redis.Client) Repository {
	return &repository{client: client}
}

func (r *repository) IncrementDailyRent(ctx context.Context, date string) error {
	key := fmt.Sprintf("stats:daily:%s", date)
	return r.client.Incr(ctx, key).Err()
}

func (r *repository) IncrementActiveRents(ctx context.Context) error {
	return r.client.Incr(ctx, "stats:active_rents").Err()
}

func (r *repository) DecrementActiveRents(ctx context.Context) error {
	return r.client.Decr(ctx, "stats:active_rents").Err()
}

func (r *repository) IncrementLocationRent(ctx context.Context, date string, location string) error {
	key := fmt.Sprintf("stats:locations:%s", date)
	return r.client.HIncrBy(ctx, key, location, 1).Err()
}

func (r *repository) GetDailyStats(ctx context.Context, date string) (int64, error) {
	key := fmt.Sprintf("stats:daily:%s", date)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func (r *repository) GetActiveRents(ctx context.Context) (int64, error) {
	val, err := r.client.Get(ctx, "stats:active_rents").Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

func (r *repository) GetLocationStats(ctx context.Context, date string) (map[string]int64, error) {
	key := fmt.Sprintf("stats:locations:%s", date)
	result := make(map[string]int64)
	
	vals, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	for k, v := range vals {
		count, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		result[k] = count
	}
	
	return result, nil
}

