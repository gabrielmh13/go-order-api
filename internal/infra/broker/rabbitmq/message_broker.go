package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"go-order-api/internal/domain/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrNotConnected = errors.New("not connected to rabbitmq")
)

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  amqp.Table
}

type MessageBroker struct {
	url     string
	conn    *amqp.Connection
	logger  logger.Logger
	mu      sync.RWMutex
	closing bool
	queues  []QueueConfig
}

func NewMessageBroker(url string, l logger.Logger, queues []QueueConfig) (*MessageBroker, error) {
	broker := &MessageBroker{
		url:    url,
		logger: l,
		queues: queues,
	}

	if err := broker.connect(); err != nil {
		return nil, err
	}

	go broker.handleReconnect()

	return broker, nil
}

func (b *MessageBroker) connect() error {
	conn, err := NewConnection(b.url, b.logger)
	if err != nil {
		return err
	}

	if err := b.declareQueues(conn); err != nil {
		conn.Close()
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.conn = conn

	return nil
}

func (b *MessageBroker) declareQueues(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	for _, q := range b.queues {
		_, err := ch.QueueDeclare(
			q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			q.NoWait,
			q.Arguments,
		)
		if err != nil {
			b.logger.Error("Failed to declare RabbitMQ queue", err, logger.Any("queue", q.Name))
			return err
		}
		b.logger.Info("Queue declared successfully", logger.Any("queue", q.Name))
	}

	return nil
}

func (b *MessageBroker) handleReconnect() {
	for {
		b.mu.RLock()
		if b.closing {
			b.mu.RUnlock()
			return
		}

		connClosed := b.conn == nil || b.conn.IsClosed()
		b.mu.RUnlock()

		if connClosed {
			b.logger.Warning("RabbitMQ connection lost, attempting to reconnect...")

			for {
				if err := b.connect(); err != nil {
					b.logger.Warning("Failed to reconnect to RabbitMQ, retrying in 3s...", logger.Any("error", err.Error()))
					time.Sleep(3 * time.Second)
				} else {
					b.logger.Info("Successfully reconnected to RabbitMQ")
					break
				}
			}
		} else {
			b.mu.RLock()
			notifyClose := b.conn.NotifyClose(make(chan *amqp.Error))
			b.mu.RUnlock()

			select {
			case err := <-notifyClose:
				if err != nil {
					b.logger.Error("RabbitMQ connection closed with error", err)
				}
			case <-time.After(5 * time.Second):
				continue
			}
		}
	}
}

func (b *MessageBroker) Publish(ctx context.Context, queue string, body any) error {
	b.mu.RLock()
	if b.conn == nil || b.conn.IsClosed() {
		b.mu.RUnlock()
		return ErrNotConnected
	}
	conn := b.conn
	b.mu.RUnlock()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	messageBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	return ch.PublishWithContext(ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})
}

func (b *MessageBroker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closing = true
	if b.conn != nil {
		b.conn.Close()
	}
}

func (b *MessageBroker) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.conn != nil && !b.conn.IsClosed()
}
