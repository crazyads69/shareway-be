package worker

import (
	"encoding/json"
	"log"
	"time"

	"shareway/infra/rabbitmq"
	"shareway/infra/ws"
	"shareway/schemas"
	"shareway/util"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WebSocketWorker struct {
	rabbitMQ *rabbitmq.RabbitMQ
	hub      *ws.Hub
	cfg      util.Config
}

func NewWebSocketWorker(rabbitMQ *rabbitmq.RabbitMQ, hub *ws.Hub, cfg util.Config) *WebSocketWorker {
	return &WebSocketWorker{
		rabbitMQ: rabbitMQ,
		hub:      hub,
		cfg:      cfg,
	}
}

func (w *WebSocketWorker) Start() {
	ch := w.rabbitMQ.GetChannel()

	// Declare queue (this will now use the passive declare first)
	// err := w.rabbitMQ.DeclareQueues()
	// if err != nil {
	// 	log.Fatalf("Failed to declare queues: %v", err)
	// }

	msgs, err := ch.Consume(
		w.cfg.AmqpWebSocketQueue,
		"",    // consumer
		false, // auto-ackddee
		false, // exclusive
		false, // no-local
		false, // no-wait
		amqp.Table{
			"x-consumer-timeout": 300000, // 5 minutes in milliseconds
		},
	)
	if err != nil {
		log.Fatalf("Failed to register a WebSocket consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var wsMessage schemas.WebSocketMessage
			err := json.Unmarshal(d.Body, &wsMessage)
			if err != nil {
				log.Printf("Error unmarshalling WebSocket message: %v", err)
				d.Reject(false)
				continue
			}

			retryCount := 0
			if d.Headers != nil {
				if count, ok := d.Headers["x-retry-count"].(int); ok {
					retryCount = count
				}
			}

			// Try to send the message to the WebSocket client
			err = w.hub.SendToUser(wsMessage.UserID, wsMessage.Type, wsMessage.Payload)
			if err != nil {
				log.Printf("Error sending WebSocket message: %v", err)

				if retryCount < w.cfg.MaxWebSocketRetries {
					if d.Headers == nil {
						d.Headers = make(amqp.Table)
					}
					d.Headers["x-retry-count"] = retryCount + 1

					// Exponential backoff
					time.Sleep(time.Second * time.Duration(1<<retryCount))
					d.Reject(true) // Reject and requeue
				} else {
					log.Printf("Max retries reached for WebSocket message to user: %s", wsMessage.UserID)
					d.Reject(false) // Send to DLQ
				}
				continue
			}

			log.Printf("WebSocket message sent successfully to user: %s", wsMessage.UserID)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] WebSocket worker waiting for messages. To exit press CTRL+C")
	<-forever
}
