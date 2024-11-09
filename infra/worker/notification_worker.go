// infra/worker/notification_worker.go
package worker

import (
	"context"
	"encoding/json"
	"log"
	"shareway/infra/fcm"
	"shareway/infra/rabbitmq"
	"shareway/schemas"
	"shareway/util"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type NotificationWorker struct {
	rabbitMQ *rabbitmq.RabbitMQ
	fcm      *fcm.FCMClient
	cfg      util.Config
}

func NewNotificationWorker(rabbitMQ *rabbitmq.RabbitMQ, fcm *fcm.FCMClient, cfg util.Config) *NotificationWorker {
	return &NotificationWorker{
		rabbitMQ: rabbitMQ,
		fcm:      fcm,
	}
}

func (nw *NotificationWorker) Start() {
	ch := nw.rabbitMQ.GetChannel()

	// Declare queue (this will now use the passive declare first)
	// err := nw.rabbitMQ.DeclareQueues()
	// if err != nil {
	// 	log.Fatalf("Failed to declare queues: %v", err)
	// }

	msgs, err := ch.Consume(
		nw.cfg.AmqpNotificationQueue,
		"",    // consumer
		false, // auto-ack (changed to false for manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		amqp.Table{
			"x-consumer-timeout": 300000, // 5 minutes in milliseconds
		},
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var notification schemas.Notification
			err := json.Unmarshal(d.Body, &notification)
			if err != nil {
				log.Printf("Error unmarshalling notification: %v", err)
				d.Reject(false) // Reject and don't requeue if unmarshal fails
				continue
			}

			// Get retry count from headers
			retryCount := 0
			if d.Headers != nil {
				if count, ok := d.Headers["x-retry-count"].(int); ok {
					retryCount = count
				}
			}

			err = nw.fcm.SendNotification(context.Background(), notification)
			if err != nil {
				log.Printf("Error sending notification: %v", err)

				if retryCount < nw.cfg.MaxNotificationRetries {
					// Increment retry count and requeue
					if d.Headers == nil {
						d.Headers = make(amqp.Table)
					}
					d.Headers["x-retry-count"] = retryCount + 1

					// Exponential backoff
					time.Sleep(time.Second * time.Duration(1<<retryCount))
					d.Reject(true) // Reject and requeue
				} else {
					log.Printf("Max retries reached for notification to token: %s", notification.Token)
					d.Reject(false) // Reject and don't requeue (will go to DLQ)
				}
				continue
			}

			log.Printf("Notification sent successfully to token: %s", notification.Token)
			d.Ack(false) // Acknowledge successful processing
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
