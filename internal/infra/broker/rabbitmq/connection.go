package rabbitmq

import (
	"go-order-api/internal/domain/logger"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewConnection(url string, l logger.Logger) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error

	for i := range 10 {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		l.Warning("RabbitMQ is not ready, retrying...", logger.Any("attempt", i+1))
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		l.Error("Failed to connect to RabbitMQ", err, logger.Any("url", url))
		return nil, err
	}

	l.Info("Connected to RabbitMQ successfully")
	return conn, nil
}
