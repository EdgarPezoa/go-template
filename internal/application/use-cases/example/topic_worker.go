package example

import (
	"context"
	"errors"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type TopicWorkerUseCase struct {
	listener ports.MessageListenerPort
	logger   ports.LoggerPort
}

func NewTopicWorkerUseCase(serviceBus ports.ServiceBusPort, destination, subscription string, logger ports.LoggerPort) (*TopicWorkerUseCase, error) {
	listener, err := serviceBus.CreateSubscriptionListener(destination, subscription)
	if err != nil {
		return nil, err
	}
	return &TopicWorkerUseCase{listener: listener, logger: logger}, nil
}

func (worker *TopicWorkerUseCase) Execute(ctx context.Context) error {
	err := worker.listener.Listen(ctx, worker.process, worker.onError)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (worker *TopicWorkerUseCase) process(ctx context.Context, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	worker.logger.Log(ctx, "Worker working")
	return nil
}

func (worker *TopicWorkerUseCase) onError(ctx context.Context, err error) {
	if !errors.Is(err, context.Canceled) {
		worker.logger.Log(ctx, "service bus listener error: "+err.Error())
	}
}
