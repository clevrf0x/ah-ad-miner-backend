package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"com.activehacks.ad-miner-backend/internal/env"
	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
}

func NewClient(redisAddr, password string) *Client {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: password,
	}

	return &Client{
		client: asynq.NewClient(redisOpt),
	}
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) EnqueueBloodhoundAnalysis(payload BloodhoundTaskPayload) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeBloodhoundAnalysis, payloadBytes)

	maxRetry := env.GetInt("WORKER_MAX_RETRIES", 3)
	maxTimeout := env.GetInt("WORKER_MAX_TIMEOUT", 60)
	opts := []asynq.Option{
		asynq.MaxRetry(maxRetry),
		asynq.Timeout(time.Duration(maxTimeout) * time.Minute),
		asynq.Queue("bloodhound"),
	}

	return c.client.Enqueue(task, opts...)
}
