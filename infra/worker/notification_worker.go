package worker

import (
	"context"
	"encoding/json"
	"log"
	"shareway/infra/fcm"
	rabbitmq "shareway/infra/rabbit_mq"
	"shareway/schemas"
	"shareway/util"
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
	msgs, err := ch.Consume(nw.cfg.AmqpNotificationQueue, // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// Consume messages from the queue
	// This is a blocking call that will run forever
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var notification schemas.Notification
			err := json.Unmarshal(d.Body, &notification)
			if err != nil {
				log.Printf("Error unmarshalling notification: %v", err)
				continue
			}

			err = nw.fcm.SendNotification(context.Background(), notification.Token, notification.Title, notification.Body)
			if err != nil {
				log.Printf("Error sending notification: %v", err)
			} else {
				log.Printf("Notification sent successfully to token: %s", notification.Token)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
