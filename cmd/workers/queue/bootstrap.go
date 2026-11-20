package main

import (
	"context"
	"fmt"

	"github.com/edgarpezoa/go-template/internal/config"
	"github.com/edgarpezoa/go-template/internal/infrastructure/adapters/servicebus"
)

type Bootstrap struct {
	factory *servicebus.ServiceBusAdapter
}

func NewBootstrap(configuration config.Config) (*Bootstrap, error) {
	factory, err := servicebus.NewServiceBusAdapter(configuration.ServiceBus.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("service bus initialization: %w", err)
	}
	return &Bootstrap{factory: factory}, nil
}

func (bootstrap *Bootstrap) Close(ctx context.Context) error {
	if bootstrap == nil {
		return nil
	}
	return bootstrap.factory.Close(ctx)
}
