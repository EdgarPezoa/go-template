//go:build integration

package servicebus

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestServiceBusQueueIntegration(t *testing.T) {
	connection := os.Getenv("AZURE_SERVICEBUS_CONNECTION_STRING")
	if connection == "" {
		connection = "Endpoint=sb://localhost;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;UseDevelopmentEmulator=true;"
	}
	queueName := os.Getenv("AZURE_SERVICEBUS_QUEUE")
	if queueName == "" {
		queueName = "go-template-messages"
	}

	factory, err := NewServiceBusAdapter(connection)
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close(context.Background())
	publisher, err := factory.CreateQueueSender(queueName)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := factory.CreateQueueListener(queueName)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	expected := fmt.Sprintf("service-bus-integration-message-%d", time.Now().UnixNano())
	if err := publisher.Publish(ctx, expected); err != nil {
		t.Fatal(err)
	}

	received := make(chan string, 1)
	listenerErrors := make(chan error, 1)
	go func() {
		listenerErrors <- listener.Listen(ctx, func(_ context.Context, message string) error {
			received <- message
			return nil
		}, nil)
	}()

	select {
	case message := <-received:
		if message != expected {
			t.Fatalf("message = %q, want %q", message, expected)
		}
	case err := <-listenerErrors:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

func TestServiceBusTopicSubscriptionIntegration(t *testing.T) {
	connection := os.Getenv("AZURE_SERVICEBUS_CONNECTION_STRING")
	if connection == "" {
		connection = "Endpoint=sb://localhost;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;UseDevelopmentEmulator=true;"
	}
	topic := os.Getenv("AZURE_SERVICEBUS_TOPIC")
	if topic == "" {
		topic = "go-template-events"
	}
	subscription := os.Getenv("AZURE_SERVICEBUS_SUBSCRIPTION")
	if subscription == "" {
		subscription = "go-template-worker"
	}
	factory, err := NewServiceBusAdapter(connection)
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close(context.Background())
	publisher, err := factory.CreateTopicSender(topic)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := factory.CreateSubscriptionListener(topic, subscription)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	expected := fmt.Sprintf("service-bus-topic-integration-message-%d", time.Now().UnixNano())
	if err := publisher.Publish(ctx, expected); err != nil {
		t.Fatal(err)
	}
	received := make(chan string, 1)
	listenerErrors := make(chan error, 1)
	go func() {
		listenerErrors <- listener.Listen(ctx, func(_ context.Context, message string) error { received <- message; return nil }, nil)
	}()
	select {
	case message := <-received:
		if message != expected {
			t.Fatalf("message = %q, want %q", message, expected)
		}
	case err := <-listenerErrors:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
