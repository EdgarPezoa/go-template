package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgarpezoa/go-template/internal/application/use-cases/example"
	"github.com/edgarpezoa/go-template/internal/config"
	loggeradapter "github.com/edgarpezoa/go-template/internal/infrastructure/adapters/logger"
)

func main() {
	logger := log.New(os.Stdout, "", 0)
	configuration, err := config.Load()
	if err != nil {
		logger.Printf("configuration failed: %v", err)
		os.Exit(1)
	}
	bootstrap, err := NewBootstrap(configuration)
	if err != nil {
		logger.Printf("worker initialization failed: %v", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	worker, err := example.NewTopicWorkerUseCase(bootstrap.factory, configuration.ServiceBus.Topic, configuration.ServiceBus.Subscription, loggeradapter.NewLoggerAdapter(logger))
	if err != nil {
		logger.Printf("worker initialization failed: %v", err)
		_ = bootstrap.Close(context.Background())
		os.Exit(1)
	}
	if err := worker.Execute(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Printf("worker execution failed: %v", err)
		_ = bootstrap.Close(context.Background())
		os.Exit(1)
	}
	if err := bootstrap.Close(context.Background()); err != nil {
		logger.Printf("worker cleanup failed: %v", err)
	}
}
