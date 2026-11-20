package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type SendQueueUseCase struct {
	sender ports.MessagePublisherPort
}

func NewSendQueueUseCase(serviceBus ports.ServiceBusPort, destination string) (*SendQueueUseCase, error) {
	sender, err := serviceBus.CreateQueueSender(destination)
	if err != nil {
		return nil, err
	}
	return &SendQueueUseCase{sender: sender}, nil
}

func (useCase *SendQueueUseCase) Execute(ctx context.Context, message string) error {
	return useCase.sender.Publish(ctx, message)
}
