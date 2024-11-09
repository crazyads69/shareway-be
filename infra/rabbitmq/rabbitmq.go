// infra/rabbitmq/rabbitmq.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"shareway/schemas"
	"shareway/util"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  util.Config
}

// NewRabbitMQ creates a new RabbitMQ instance

// NewRabbitMQ creates a new RabbitMQ instance
func NewRabbitMQ(cfg util.Config) (*RabbitMQ, error) {
	// Config connection and heartbeat
	conn, err := amqp.DialConfig(cfg.AmqpServerURL, amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Set QoS for better load distribution
	// err = ch.Qos(
	// 	1,     // prefetch count
	// 	0,     // prefetch size
	// 	false, // global
	// )
	// if err != nil {
	// 	ch.Close()
	// 	conn.Close()
	// 	return nil, fmt.Errorf("failed to set QoS: %w", err)
	// }

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

func ConnectRabbitMQ(cfg util.Config) (*RabbitMQ, error) {
	var rabbitMQ *RabbitMQ
	var err error
	for i := 0; i < 5; i++ { // Try 5 times
		rabbitMQ, err = NewRabbitMQ(cfg)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", i+1, err)
		time.Sleep(time.Second * 5) // Wait 5 seconds before retrying
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after 5 attempts: %v", err)
	}
	return rabbitMQ, nil
}

func (r *RabbitMQ) DeclareQueues() error {
	// Declare notification queue
	err := r.declareQueueIfNotExists(r.config.AmqpNotificationQueue, "notification")
	if err != nil {
		return err
	}

	// Declare WebSocket queue
	err = r.declareQueueIfNotExists(r.config.AmqpWebSocketQueue, "websocket")
	if err != nil {
		return err
	}

	return nil
}

// func (r *RabbitMQ) declareQueueIfNotExists(queueName, queueType string) error {
// 	// Try to declare the queue passively (check if it exists)
// 	_, err := r.channel.QueueDeclarePassive(
// 		queueName,
// 		true,  // durable
// 		false, // delete when unused
// 		false, // exclusive
// 		false, // no-wait
// 		nil,   // arguments
// 	)

// 	if err != nil {
// 		// If the queue doesn't exist, declare it
// 		dlxName := queueType + ".dlx"
// 		dlqName := queueType + ".dlq"

// 		args := amqp.Table{
// 			"x-dead-letter-exchange":    queueType + ".dlx",
// 			"x-dead-letter-routing-key": queueType + ".dlq",
// 			"x-message-ttl":             300000, // 5 minutes
// 			"x-consumer-timeout":        300000,
// 		}

// 		_, err = r.channel.QueueDeclare(
// 			queueName,
// 			true,  // durable
// 			false, // delete when unused
// 			false, // exclusive
// 			false, // no-wait
// 			args,  // arguments
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to declare %s queue: %w", queueType, err)
// 		}

// 		// Declare the dead-letter exchange
// 		err = r.channel.ExchangeDeclare(
// 			dlxName,  // name
// 			"direct", // type
// 			true,     // durable
// 			false,    // auto-deleted
// 			false,    // internal
// 			false,    // no-wait
// 			nil,      // arguments
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to declare %s DLX: %w", queueType, err)
// 		}

// 		// Declare the dead-letter queue
// 		_, err = r.channel.QueueDeclare(
// 			dlqName, // name
// 			true,    // durable
// 			false,   // delete when unused
// 			false,   // exclusive
// 			false,   // no-wait
// 			nil,     // arguments
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to declare %s DLQ: %w", queueType, err)
// 		}

// 		// Bind the DLQ to the DLX
// 		err = r.channel.QueueBind(
// 			dlqName, // queue name
// 			dlqName, // routing key
// 			dlxName, // exchange
// 			false,
// 			nil,
// 		)
// 		if err != nil {
// 			return fmt.Errorf("failed to bind %s DLQ to DLX: %w", queueType, err)
// 		}
// 	}

// 	return nil
// }

func (r *RabbitMQ) declareQueueIfNotExists(queueName, queueType string) error {
	// Try to declare the queue (this will create it if it doesn't exist)
	dlxName := queueType + ".dlx"
	dlqName := queueType + ".dlq"

	args := amqp.Table{
		"x-dead-letter-exchange":    dlxName,
		"x-dead-letter-routing-key": dlqName,
		"x-message-ttl":             300000, // 5 minutes
		"x-consumer-timeout":        300000,
	}

	_, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare %s queue: %w", queueType, err)
	}

	// Declare the dead-letter exchange
	err = r.channel.ExchangeDeclare(
		dlxName,  // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare %s DLX: %w", queueType, err)
	}

	// Declare the dead-letter queue
	_, err = r.channel.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare %s DLQ: %w", queueType, err)
	}

	// Bind the DLQ to the DLX
	err = r.channel.QueueBind(
		dlqName, // queue name
		dlqName, // routing key
		dlxName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind %s DLQ to DLX: %w", queueType, err)
	}

	return nil
}

// PublishWebSocketMessage publishes a message to the WebSocket queue
func (r *RabbitMQ) PublishWebSocketMessage(ctx context.Context, message schemas.WebSocketMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal WebSocket message: %w", err)
	}

	// Enable publisher confirms
	err = r.channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("failed to put channel in confirm mode: %w", err)
	}

	confirms := r.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	err = r.channel.PublishWithContext(ctx,
		"",                          // exchange
		r.config.AmqpWebSocketQueue, // routing key
		true,                        // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish WebSocket message: %w", err)
	}

	// Wait for confirmation
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("failed to receive publish confirmation")
		}
	case <-ctx.Done():
		return fmt.Errorf("publish confirmation timeout: %w", ctx.Err())
	}

	return nil
}

// Close closes the channel and connection
func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

// GetChannel returns the amqp.Channel
func (r *RabbitMQ) GetChannel() *amqp.Channel {
	return r.channel
}

// PublishNotification publishes a notification to the queue for notifications
func (r *RabbitMQ) PublishNotification(ctx context.Context, notification schemas.Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Enable publisher confirms
	err = r.channel.Confirm(false)
	if err != nil {
		return fmt.Errorf("failed to put channel in confirm mode: %w", err)
	}

	confirms := r.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	err = r.channel.PublishWithContext(ctx,
		"",                             // exchange
		r.config.AmqpNotificationQueue, // routing key
		true,                           // mandatory
		false,                          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	// Wait for confirmation
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("failed to receive publish confirmation")
		}
	case <-ctx.Done():
		return fmt.Errorf("publish confirmation timeout: %w", ctx.Err())
	}

	return nil
}
