package worker

import (
	"encoding/json"
	"log"
	"shareway/infra/rabbitmq"
	"shareway/infra/ws"
	"shareway/util"
	"time"

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

type WebSocketMessage struct {
	UserID  string      `json:"user_id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func (w *WebSocketWorker) Start() {
	ch := w.rabbitMQ.GetChannel()

	// Declare WebSocket message queue
	queueName := w.cfg.AmqpWebSocketQueue
	_, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "websocket.dlx",
			"x-dead-letter-routing-key": "websocket.dlq",
		},
	)
	if err != nil {
		log.Fatalf("Failed to declare WebSocket queue: %v", err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Failed to register a WebSocket consumer: %v", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var wsMessage WebSocketMessage
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
