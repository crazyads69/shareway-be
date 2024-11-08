package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
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
	conn, err := amqp.Dial(cfg.AmqpServerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Set QoS for better load distribution
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

// DeclareQueues declares all required queues
func (r *RabbitMQ) DeclareQueues() error {
	// Declare notification queue
	err := r.declareNotificationQueue()
	if err != nil {
		return err
	}

	// Declare WebSocket queue
	err = r.declareWebSocketQueue()
	if err != nil {
		return err
	}

	return nil
}

// declareNotificationQueue declares the queue for notifications
func (r *RabbitMQ) declareNotificationQueue() error {
	// Declare the main queue
	_, err := r.channel.QueueDeclare(
		r.config.AmqpNotificationQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    "notification.dlx",
			"x-dead-letter-routing-key": "notification.dlq",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare main queue: %w", err)
	}

	// Declare the dead-letter exchange
	err = r.channel.ExchangeDeclare(
		"notification.dlx", // name
		"direct",           // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLX: %w", err)
	}

	// Declare the dead-letter queue
	_, err = r.channel.QueueDeclare(
		"notification.dlq", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Bind the DLQ to the DLX
	err = r.channel.QueueBind(
		"notification.dlq", // queue name
		"notification.dlq", // routing key
		"notification.dlx", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ to DLX: %w", err)
	}

	return nil
}

// declareWebSocketQueue declares the queue for WebSocket messages
func (r *RabbitMQ) declareWebSocketQueue() error {
	// Declare the main WebSocket queue
	_, err := r.channel.QueueDeclare(
		r.config.AmqpWebSocketQueue,
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
		return fmt.Errorf("failed to declare WebSocket queue: %w", err)
	}

	// Declare the dead-letter exchange for WebSocket
	err = r.channel.ExchangeDeclare(
		"websocket.dlx", // name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare WebSocket DLX: %w", err)
	}

	// Declare the dead-letter queue for WebSocket
	_, err = r.channel.QueueDeclare(
		"websocket.dlq", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare WebSocket DLQ: %w", err)
	}

	// Bind the WebSocket DLQ to the DLX
	err = r.channel.QueueBind(
		"websocket.dlq", // queue name
		"websocket.dlq", // routing key
		"websocket.dlx", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind WebSocket DLQ to DLX: %w", err)
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
