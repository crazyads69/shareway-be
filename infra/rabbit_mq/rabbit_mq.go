package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"shareway/schemas"
	"shareway/util"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  util.Config
}

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

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
		config:  cfg,
	}, nil
}

// DeclareQueue declares a queue for notifications
func (r *RabbitMQ) DeclareQueue() error {
	_, err := r.channel.QueueDeclare(
		r.config.AmqpNotificationQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
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

	err = r.channel.PublishWithContext(ctx,
		"",                             // exchange
		r.config.AmqpNotificationQueue, // routing key
		false,                          // mandatory
		false,                          // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}
