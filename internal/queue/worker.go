package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Worker struct {
	server *asynq.Server
}

func NewWorker(redisAddr, password string, logger *slog.Logger) *Worker {
	redisOpt := asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: password,
	}

	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 1, // Process one task at a time
		Queues: map[string]int{
			"bloodhound": 1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Error(fmt.Sprintf("Task failed: %s", task.Type()), "Error", err)
		}),
	})

	return &Worker{
		server: server,
	}
}

func (w *Worker) Start(handler *TaskHandler, logger *slog.Logger) error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeBloodhoundAnalysis, handler.ProcessBloodhoundAnalysis)

	logger.Info("Starting Asynq Worker")
	return w.server.Run(mux)
}

func (w *Worker) Shutdown() {
	w.server.Shutdown()
}
