package servicebus

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type listener struct {
	receiver receiver
}

type receivedMessage struct {
	body    string
	message *azservicebus.ReceivedMessage
}

type receiver interface {
	Receive(context.Context) ([]receivedMessage, error)
	Complete(context.Context, receivedMessage) error
	Abandon(context.Context, receivedMessage) error
	Close(context.Context) error
}

type azureReceiver struct {
	receiver *azservicebus.Receiver
}

func (receiver *azureReceiver) Receive(ctx context.Context) ([]receivedMessage, error) {
	messages, err := receiver.receiver.ReceiveMessages(ctx, 1, nil)
	if err != nil {
		return nil, err
	}
	result := make([]receivedMessage, len(messages))
	for index, message := range messages {
		result[index] = receivedMessage{body: string(message.Body), message: message}
	}
	return result, nil
}

func (receiver *azureReceiver) Complete(ctx context.Context, message receivedMessage) error {
	return receiver.receiver.CompleteMessage(ctx, message.message, nil)
}

func (receiver *azureReceiver) Abandon(ctx context.Context, message receivedMessage) error {
	return receiver.receiver.AbandonMessage(ctx, message.message, nil)
}

func (receiver *azureReceiver) Close(ctx context.Context) error {
	return receiver.receiver.Close(ctx)
}

func (listener *listener) Listen(
	ctx context.Context,
	onProcess ports.OnProcess,
	onError ports.OnError,
) error {
	for {
		messages, err := listener.receiver.Receive(ctx)
		if err != nil {
			if onError != nil {
				onError(ctx, err)
			}
			return err
		}
		for _, message := range messages {
			if err := onProcess(ctx, message.body); err != nil {
				if onError != nil {
					onError(ctx, err)
				}
				if err := listener.receiver.Abandon(ctx, message); err != nil {
					if onError != nil {
						onError(ctx, err)
					}
					return err
				}
				continue
			}
			if err := listener.receiver.Complete(ctx, message); err != nil {
				if onError != nil {
					onError(ctx, err)
				}
				return err
			}
		}
	}
}

func (listener *listener) Close(ctx context.Context) error {
	if listener == nil {
		return nil
	}
	if err := listener.receiver.Close(ctx); err != nil {
		return err
	}
	return nil
}
