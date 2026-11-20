package example

import (
	"context"
	"errors"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type QueueWorkerUseCase struct {
	listener ports.MessageListenerPort
	logger   ports.LoggerPort
}

func NewQueueWorkerUseCase(serviceBus ports.ServiceBusPort, destination string, logger ports.LoggerPort) (*QueueWorkerUseCase, error) {
	listener, err := serviceBus.CreateQueueListener(destination)
	if err != nil {
		return nil, err
	}
	return &QueueWorkerUseCase{listener: listener, logger: logger}, nil
}

func (worker *QueueWorkerUseCase) Execute(ctx context.Context) error {
	err := worker.listener.Listen(ctx, worker.process, worker.onError)
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (worker *QueueWorkerUseCase) process(ctx context.Context, _ string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	worker.logger.Log(ctx, "Worker working")
	return nil
}

func (worker *QueueWorkerUseCase) onError(ctx context.Context, err error) {
	if !errors.Is(err, context.Canceled) {
		worker.logger.Log(ctx, "service bus listener error: "+err.Error())
	}
}
