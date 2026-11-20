package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type ExampleWorkerUseCase struct {
	logger ports.LoggerPort
}

func NewExampleWorkerUseCase(logger ports.LoggerPort) *ExampleWorkerUseCase {
	return &ExampleWorkerUseCase{logger: logger}
}

func (worker *ExampleWorkerUseCase) Execute(ctx context.Context) error {
	worker.logger.Log(ctx, "Worker working")
	return nil
}
