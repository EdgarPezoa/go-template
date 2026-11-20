package servicebus

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type publisher struct {
	sender sender
}

type sender interface {
	Send(context.Context, string) error
	Close(context.Context) error
}

type azureSender struct {
	sender *azservicebus.Sender
}

func (sender *azureSender) Send(ctx context.Context, message string) error {
	return sender.sender.SendMessage(ctx, &azservicebus.Message{Body: []byte(message)}, nil)
}

func (sender *azureSender) Close(ctx context.Context) error {
	return sender.sender.Close(ctx)
}

func (publisher *publisher) Publish(ctx context.Context, message string) error {
	return publisher.sender.Send(ctx, message)
}

func (publisher *publisher) Close(ctx context.Context) error {
	if publisher == nil {
		return nil
	}
	if err := publisher.sender.Close(ctx); err != nil {
		return err
	}
	return nil
}
