package task

import (
	"fmt"
	"shareway/util"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

type AsyncServer struct {
	AsynqServer *asynq.Server
}

func NewAsynqServer(cfg util.Config) *AsyncServer {
	redisAddr := fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr,
			DB: 1}, // Default DB is 0 use for caching so we use DB 1
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)
	return &AsyncServer{
		AsynqServer: srv,
	}
}

// StartAsynqServer starts the asynq server in a goroutine
func (as *AsyncServer) StartAsynqServer(processor *TaskProcessor) {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeWebsocketMessage, processor.HandleWebsocketMessageTask)
	mux.HandleFunc(TypeFCMNofitication, processor.HandleFCMNotificationTask)

	// Start the server in a goroutine
	go func() {
		if err := as.AsynqServer.Run(mux); err != nil {
			log.Error().Err(err).Msg("Failed to start Asynq server")
		}
	}()

	log.Info().Msg("Asynq server started successfully")
}

// Shutdown gracefully shuts down the Asynq server
func (as *AsyncServer) Shutdown() {
	as.AsynqServer.Stop()
	log.Info().Msg("Asynq server stopped")
}
