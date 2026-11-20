package ports

import "context"

type MessagePublisherPort interface {
	Publish(context.Context, string) error
	Close(context.Context) error
}
type OnProcess func(context.Context, string) error
type OnError func(context.Context, error)
type MessageListenerPort interface {
	Listen(context.Context, OnProcess, OnError) error
	Close(context.Context) error
}

type ServiceBusPort interface {
	CreateQueueSender(string) (MessagePublisherPort, error)
	CreateTopicSender(string) (MessagePublisherPort, error)
	CreateQueueListener(string) (MessageListenerPort, error)
	CreateSubscriptionListener(string, string) (MessageListenerPort, error)
}
