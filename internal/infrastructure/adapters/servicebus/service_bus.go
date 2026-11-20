package servicebus

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type ServiceBusAdapter struct {
	client serviceBusClient
	links  []func(context.Context) error
}

type serviceBusClient interface {
	NewSender(string) (sender, error)
	NewQueueReceiver(string) (receiver, error)
	NewSubscriptionReceiver(string, string) (receiver, error)
	Close(context.Context) error
}

type azureServiceBusClient struct {
	client *azservicebus.Client
}

var newServiceBusClient = func(connectionString string) (serviceBusClient, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}
	return &azureServiceBusClient{client: client}, nil
}

func (client *azureServiceBusClient) NewSender(name string) (sender, error) {
	value, err := client.client.NewSender(name, nil)
	if err != nil {
		return nil, err
	}
	return &azureSender{sender: value}, nil
}

func (client *azureServiceBusClient) NewQueueReceiver(queue string) (receiver, error) {
	value, err := client.client.NewReceiverForQueue(queue, nil)
	if err != nil {
		return nil, err
	}
	return &azureReceiver{receiver: value}, nil
}

func (client *azureServiceBusClient) NewSubscriptionReceiver(topic, subscription string) (receiver, error) {
	value, err := client.client.NewReceiverForSubscription(topic, subscription, nil)
	if err != nil {
		return nil, err
	}
	return &azureReceiver{receiver: value}, nil
}

func (client *azureServiceBusClient) Close(ctx context.Context) error {
	return client.client.Close(ctx)
}

func NewServiceBusAdapter(connectionString string) (*ServiceBusAdapter, error) {
	if connectionString == "" {
		return nil, errors.New("service bus connection string is required")
	}
	client, err := newServiceBusClient(connectionString)
	if err != nil {
		return nil, err
	}
	return &ServiceBusAdapter{client: client}, nil
}

func (serviceBus *ServiceBusAdapter) CreateQueueSender(queue string) (ports.MessagePublisherPort, error) {
	if queue == "" {
		return nil, errors.New("service bus queue is required")
	}
	sender, err := serviceBus.client.NewSender(queue)
	if err != nil {
		return nil, err
	}
	serviceBus.links = append(serviceBus.links, sender.Close)
	return &publisher{sender: sender}, nil
}

func (serviceBus *ServiceBusAdapter) CreateTopicSender(topic string) (ports.MessagePublisherPort, error) {
	if topic == "" {
		return nil, errors.New("service bus topic is required")
	}
	sender, err := serviceBus.client.NewSender(topic)
	if err != nil {
		return nil, err
	}
	serviceBus.links = append(serviceBus.links, sender.Close)
	return &publisher{sender: sender}, nil
}

func (serviceBus *ServiceBusAdapter) CreateQueueListener(queue string) (ports.MessageListenerPort, error) {
	if queue == "" {
		return nil, errors.New("service bus queue is required")
	}
	receiver, err := serviceBus.client.NewQueueReceiver(queue)
	if err != nil {
		return nil, err
	}
	serviceBus.links = append(serviceBus.links, receiver.Close)
	return &listener{receiver: receiver}, nil
}

func (serviceBus *ServiceBusAdapter) CreateSubscriptionListener(topic, subscription string) (ports.MessageListenerPort, error) {
	if topic == "" || subscription == "" {
		return nil, errors.New("service bus topic and subscription are required")
	}
	receiver, err := serviceBus.client.NewSubscriptionReceiver(topic, subscription)
	if err != nil {
		return nil, err
	}
	serviceBus.links = append(serviceBus.links, receiver.Close)
	return &listener{receiver: receiver}, nil
}

func (serviceBus *ServiceBusAdapter) Close(ctx context.Context) error {
	if serviceBus == nil {
		return nil
	}
	var closeErr error
	for index := len(serviceBus.links) - 1; index >= 0; index-- {
		if err := serviceBus.links[index](ctx); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	if err := serviceBus.client.Close(ctx); err != nil && closeErr == nil {
		closeErr = err
	}
	return closeErr
}
