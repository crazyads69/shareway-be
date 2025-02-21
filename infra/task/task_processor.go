package task

import (
	"context"
	"encoding/json"
	"log"

	"shareway/infra/fcm"
	"shareway/infra/ws"
	"shareway/schemas"
	"shareway/util"

	"github.com/hibiken/asynq"
)

type TaskProcessor struct {
	hub       *ws.Hub
	cfg       util.Config
	fcmClient *fcm.FCMClient
}

func NewTaskProcessor(hub *ws.Hub, cfg util.Config, fcmClient *fcm.FCMClient) *TaskProcessor {
	return &TaskProcessor{
		hub:       hub,
		cfg:       cfg,
		fcmClient: fcmClient,
	}
}

// Handle websocket message task
func (tp *TaskProcessor) HandleWebsocketMessageTask(ctx context.Context, t *asynq.Task) error {
	var wsMessage schemas.WebSocketMessage
	if err := json.Unmarshal(t.Payload(), &wsMessage); err != nil {
		return err
	}
	err := tp.hub.SendToUser(wsMessage.UserID, wsMessage.Type, wsMessage.Payload)
	if err != nil {
		return err
	}
	log.Printf("Sent message to user success %s: %v", wsMessage.UserID, wsMessage)
	return nil
}

// Handle FCM notification task
func (tp *TaskProcessor) HandleFCMNotificationTask(ctx context.Context, t *asynq.Task) error {
	var notification schemas.Notification
	if err := json.Unmarshal(t.Payload(), &notification); err != nil {
		return err
	}
	err := tp.fcmClient.SendNotification(ctx, notification)
	if err != nil {
		return err
	}
	log.Printf("Sent FCM notification success: %v", notification)
	return nil
}

// RegisterTasks registers all tasks that this processor can handle.
