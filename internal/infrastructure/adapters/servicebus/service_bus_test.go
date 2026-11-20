package servicebus

import (
	"context"
	"reflect"
	"testing"
)

type fakeServiceBusClient struct {
	events []string
}

func (client *fakeServiceBusClient) NewSender(string) (sender, error) {
	client.events = append(client.events, "sender")
	return &fakeSender{close: func() error {
		client.events = append(client.events, "sender-close")
		return nil
	}}, nil
}

func (client *fakeServiceBusClient) NewQueueReceiver(string) (receiver, error) {
	client.events = append(client.events, "queue-receiver")
	return &fakeReceiver{close: func() error {
		client.events = append(client.events, "queue-receiver-close")
		return nil
	}}, nil
}

func (client *fakeServiceBusClient) NewSubscriptionReceiver(string, string) (receiver, error) {
	client.events = append(client.events, "subscription-receiver")
	return &fakeReceiver{close: func() error {
		client.events = append(client.events, "subscription-receiver-close")
		return nil
	}}, nil
}

func (client *fakeServiceBusClient) Close(context.Context) error {
	client.events = append(client.events, "client-close")
	return nil
}

func TestServiceBusClosesChildrenInReverseOrderThenClient(t *testing.T) {
	client := &fakeServiceBusClient{}
	serviceBus := &ServiceBusAdapter{client: client}

	if _, err := serviceBus.CreateQueueSender("queue"); err != nil {
		t.Fatal(err)
	}
	if _, err := serviceBus.CreateQueueListener("queue"); err != nil {
		t.Fatal(err)
	}

	if err := serviceBus.Close(context.Background()); err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"sender",
		"queue-receiver",
		"queue-receiver-close",
		"sender-close",
		"client-close",
	}
	if !reflect.DeepEqual(client.events, expected) {
		t.Fatalf("events = %v, want %v", client.events, expected)
	}
}

func TestServiceBusDoesNotCloseClientDuringChildCreation(t *testing.T) {
	client := &fakeServiceBusClient{}
	serviceBus := &ServiceBusAdapter{client: client}

	if _, err := serviceBus.CreateSubscriptionListener("topic", "subscription"); err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(client.events, []string{"client-close"}) ||
		(client.events != nil && client.events[len(client.events)-1] == "client-close") {
		t.Fatalf("client closed during child creation: %v", client.events)
	}
}
