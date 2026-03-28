package broker

import "context"

type IMessageBroker interface {
	Publish(ctx context.Context, queue string, body any) error
}
