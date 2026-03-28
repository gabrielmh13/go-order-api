package rabbitmq

import (
	"context"
	"errors"
)

type StatusProvider interface {
	IsConnected() bool
}

type RabbitMQHealthChecker struct {
	broker StatusProvider
}

func NewRabbitMQHealthChecker(broker StatusProvider) *RabbitMQHealthChecker {
	return &RabbitMQHealthChecker{
		broker: broker,
	}
}

func (h *RabbitMQHealthChecker) Name() string {
	return "rabbitmq"
}

func (h *RabbitMQHealthChecker) Check(ctx context.Context) error {
	if h.broker == nil || !h.broker.IsConnected() {
		return errors.New("rabbitmq connection is closed or reestablishing")
	}
	return nil
}
