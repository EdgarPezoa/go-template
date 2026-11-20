package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgarpezoa/go-template/internal/config"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
	redisadapter "github.com/edgarpezoa/go-template/internal/infrastructure/adapters/redis"
	"github.com/edgarpezoa/go-template/internal/infrastructure/adapters/servicebus"
	mongodbadapter "github.com/edgarpezoa/go-template/internal/infrastructure/adapters/mongodb"
	examplecontroller "github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example"
	healthcontroller "github.com/edgarpezoa/go-template/internal/infrastructure/controllers/health"
)

type Bootstrap struct {
	Handler http.Handler
	close   func(context.Context) error
}

func registerControllers(redisClient *redisadapter.RedisAdapter, serviceBus ports.ServiceBusPort, database ports.DatabasePort, configuration config.ServiceBusConfig) (http.Handler, error) {
	router := http.NewServeMux()
	healthcontroller.NewHealthController(func() bool { return true }).Routes(router)
	controller, err := examplecontroller.NewExampleController(redisClient, serviceBus, database, configuration)
	if err != nil {
		return nil, err
	}
	controller.Routes(router)
	return router, nil
}

func NewBootstrap(configuration config.Config) (*Bootstrap, error) {
	redisClient, err := redisadapter.NewRedisAdapter(configuration.Redis.Address, configuration.Redis.Password, configuration.Redis.DB)
	if err != nil {
		return nil, fmt.Errorf("redis initialization: %w", err)
	}

	serviceBus, err := servicebus.NewServiceBusAdapter(configuration.ServiceBus.ConnectionString)
	if err != nil {
		_ = redisClient.Close()
		return nil, fmt.Errorf("service bus initialization: %w", err)
	}
	database, err := mongodbadapter.NewMongoDBAdapter(configuration.Mongo.URL, configuration.Mongo.User, configuration.Mongo.Password, configuration.Mongo.Args, configuration.Mongo.Database, configuration.Mongo.Port, configuration.Mongo.ServerSelectionTimeoutSeconds, configuration.Mongo.PingTimeoutSeconds)
	if err != nil {
		_ = serviceBus.Close(context.Background())
		_ = redisClient.Close()
		return nil, fmt.Errorf("mongodb initialization: %w", err)
	}

	closers := []func(context.Context) error{
		func(context.Context) error { return redisClient.Close() },
		func(ctx context.Context) error { return serviceBus.Close(ctx) },
		func(ctx context.Context) error { return database.Disconnect(ctx) },
	}

	handler, err := registerControllers(redisClient, serviceBus, database, configuration.ServiceBus)
	if err != nil {
		_ = closeResources(closers, context.Background())
		return nil, err
	}
	return &Bootstrap{
		Handler: handler,
		close: func(ctx context.Context) error {
			return closeResources(closers, ctx)
		},
	}, nil
}

func closeResources(closers []func(context.Context) error, ctx context.Context) error {
	var closeErr error
	for index := len(closers) - 1; index >= 0; index-- {
		if err := closers[index](ctx); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

func (bootstrap *Bootstrap) Close(ctx context.Context) error {
	if bootstrap == nil || bootstrap.close == nil {
		return nil
	}
	return bootstrap.close(ctx)
}
