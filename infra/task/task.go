package task

import (
	"encoding/json"
	"fmt"
	"shareway/schemas"
	"shareway/util"

	"github.com/hibiken/asynq"
)

const (
	TypeWebsocketMessage = "websocket:message"
	TypeFCMNofitication  = "notification:fcm"
)

type AsyncClient struct {
	AsynqClient *asynq.Client
}

func NewAsynqClient(cfg util.Config) *AsyncClient {
	redisAddr := fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
	cli := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr,
		DB: 1}) // Default DB is 0 use for caching so we use DB 1
	return &AsyncClient{
		AsynqClient: cli,
	}
}

// EnqueueWebsocketMessage enqueues a websocket message task
func (ac *AsyncClient) EnqueueWebsocketMessage(wsMessage schemas.WebSocketMessage) error {

	// Marshal the task payload
	bytes, err := json.Marshal(wsMessage)
	if err != nil {
		return err
	}

	// Create a new task
	task := asynq.NewTask(TypeWebsocketMessage, bytes)

	// Enqueue the task
	_, err = ac.AsynqClient.Enqueue(task,
		asynq.MaxRetry(0),
	)
	return err
}

// EnqueueFCMNotification enqueues a FCM notification task
func (ac *AsyncClient) EnqueueFCMNotification(notification schemas.Notification) error {

	// Marshal the task payload
	bytes, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	// Create a new task
	task := asynq.NewTask(TypeFCMNofitication, bytes)

	// Enqueue the task
	_, err = ac.AsynqClient.Enqueue(task,
		asynq.MaxRetry(0),
	)
	return err
}
