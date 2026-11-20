package main

import (
	"context"
	"log"
	"os"

	worker "github.com/edgarpezoa/go-template/internal/application/use-cases/example"
	loggeradapter "github.com/edgarpezoa/go-template/internal/infrastructure/adapters/logger"
)

func main() {
	logger := log.New(os.Stdout, "", 0)
	useCase := worker.NewExampleWorkerUseCase(loggeradapter.NewLoggerAdapter(logger))
	if err := useCase.Execute(context.Background()); err != nil {
		logger.Print(err)
		os.Exit(1)
	}
}
