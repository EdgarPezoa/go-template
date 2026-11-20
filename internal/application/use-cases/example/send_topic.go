package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type SendTopicUseCase struct {
	sender ports.MessagePublisherPort
}

func NewSendTopicUseCase(serviceBus ports.ServiceBusPort, destination string) (*SendTopicUseCase, error) {
	sender, err := serviceBus.CreateTopicSender(destination)
	if err != nil {
		return nil, err
	}
	return &SendTopicUseCase{sender: sender}, nil
}

func (useCase *SendTopicUseCase) Execute(ctx context.Context, message string) error {
	return useCase.sender.Publish(ctx, message)
}
