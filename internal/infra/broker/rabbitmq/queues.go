package rabbitmq

const (
	OrderStatusChangedQueue = "order.status.changed"
)

func GetDefaultQueues() []QueueConfig {
	return []QueueConfig{
		{
			Name:    OrderStatusChangedQueue,
			Durable: true,
		},
	}
}
