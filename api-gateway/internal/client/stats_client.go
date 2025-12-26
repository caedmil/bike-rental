package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type StatsClient interface {
	GetDailyStats(ctx context.Context, date string) (map[string]interface{}, error)
	GetActiveRents(ctx context.Context) (int64, error)
}

type statsClient struct {
	baseURL string
	client  *http.Client
}

func NewStatsClient(baseURL string) StatsClient {
	return &statsClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *statsClient) GetDailyStats(ctx context.Context, date string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/internal/stats/daily?date=%s", c.baseURL, date)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stats service returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *statsClient) GetActiveRents(ctx context.Context) (int64, error) {
	url := fmt.Sprintf("%s/internal/stats/active", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("stats service returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	activeRents, ok := result["active_rents"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid response format")
	}

	return int64(activeRents), nil
}

