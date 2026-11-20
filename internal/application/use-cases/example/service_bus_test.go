package example

import (
	"context"
	"testing"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type fakeSender struct {
	messages []string
}

func (sender *fakeSender) Publish(_ context.Context, message string) error {
	sender.messages = append(sender.messages, message)
	return nil
}

func (sender *fakeSender) Close(context.Context) error { return nil }

type fakeServiceBus struct {
	queueSender, topicSender *fakeSender
	queueSenders, topicSenders int
}

func (serviceBus *fakeServiceBus) CreateQueueSender(string) (ports.MessagePublisherPort, error) {
	serviceBus.queueSenders++
	serviceBus.queueSender = &fakeSender{}
	return serviceBus.queueSender, nil
}

func (serviceBus *fakeServiceBus) CreateTopicSender(string) (ports.MessagePublisherPort, error) {
	serviceBus.topicSenders++
	serviceBus.topicSender = &fakeSender{}
	return serviceBus.topicSender, nil
}

func (serviceBus *fakeServiceBus) CreateQueueListener(string) (ports.MessageListenerPort, error) {
	return nil, nil
}

func (serviceBus *fakeServiceBus) CreateSubscriptionListener(string, string) (ports.MessageListenerPort, error) {
	return nil, nil
}

func TestSendQueueUseCaseCreatesSenderOnceAndExecutes(t *testing.T) {
	serviceBus := &fakeServiceBus{}
	useCase, err := NewSendQueueUseCase(serviceBus, "queue")
	if err != nil {
		t.Fatal(err)
	}
	if err := useCase.Execute(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}
	if serviceBus.queueSenders != 1 {
		t.Fatalf("queue senders = %d, want 1", serviceBus.queueSenders)
	}
	if got := serviceBus.queueSender.messages; len(got) != 1 || got[0] != "message" {
		t.Fatalf("messages = %v, want [message]", got)
	}
}

func TestSendTopicUseCaseCreatesSenderOnceAndExecutes(t *testing.T) {
	serviceBus := &fakeServiceBus{}
	useCase, err := NewSendTopicUseCase(serviceBus, "topic")
	if err != nil {
		t.Fatal(err)
	}
	if err := useCase.Execute(context.Background(), "message"); err != nil {
		t.Fatal(err)
	}
	if serviceBus.topicSenders != 1 {
		t.Fatalf("topic senders = %d, want 1", serviceBus.topicSenders)
	}
	if got := serviceBus.topicSender.messages; len(got) != 1 || got[0] != "message" {
		t.Fatalf("messages = %v, want [message]", got)
	}
}
