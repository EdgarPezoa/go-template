package example

import (
	"bytes"
	"context"
	"testing"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type testLogger struct {
	output bytes.Buffer
}

func (logger *testLogger) Log(_ context.Context, message string) {
	logger.output.WriteString(message)
	logger.output.WriteByte('\n')
}

type fakeListener struct {
	listens int
	process ports.OnProcess
	onError ports.OnError
}

func (listener *fakeListener) Listen(_ context.Context, process ports.OnProcess, onError ports.OnError) error {
	listener.listens++
	listener.process = process
	listener.onError = onError
	return process(context.Background(), "message")
}

func (listener *fakeListener) Close(context.Context) error { return nil }

type fakeWorkerServiceBus struct {
	queueListener, topicListener *fakeListener
	queueListeners, topicListeners int
}

func (serviceBus *fakeWorkerServiceBus) CreateQueueSender(string) (ports.MessagePublisherPort, error) {
	return nil, nil
}

func (serviceBus *fakeWorkerServiceBus) CreateTopicSender(string) (ports.MessagePublisherPort, error) {
	return nil, nil
}

func (serviceBus *fakeWorkerServiceBus) CreateQueueListener(string) (ports.MessageListenerPort, error) {
	serviceBus.queueListeners++
	serviceBus.queueListener = &fakeListener{}
	return serviceBus.queueListener, nil
}

func (serviceBus *fakeWorkerServiceBus) CreateSubscriptionListener(string, string) (ports.MessageListenerPort, error) {
	serviceBus.topicListeners++
	serviceBus.topicListener = &fakeListener{}
	return serviceBus.topicListener, nil
}

func TestQueueWorkerUseCaseCreatesListenerOnceAndExecutes(t *testing.T) {
	logger := &testLogger{}
	serviceBus := &fakeWorkerServiceBus{}
	worker, err := NewQueueWorkerUseCase(serviceBus, "queue", logger)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Execute(context.Background()); err != nil {
		t.Fatal(err)
	}
	if serviceBus.queueListeners != 1 || serviceBus.queueListener.listens != 1 {
		t.Fatalf("queue listeners = %d, listens = %d, want 1, 1", serviceBus.queueListeners, serviceBus.queueListener.listens)
	}
	if got := logger.output.String(); got != "Worker working\n" {
		t.Fatalf("output = %q, want %q", got, "Worker working\n")
	}
}

func TestTopicWorkerUseCaseCreatesListenerOnceAndExecutes(t *testing.T) {
	logger := &testLogger{}
	serviceBus := &fakeWorkerServiceBus{}
	worker, err := NewTopicWorkerUseCase(serviceBus, "topic", "subscription", logger)
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Execute(context.Background()); err != nil {
		t.Fatal(err)
	}
	if serviceBus.topicListeners != 1 || serviceBus.topicListener.listens != 1 {
		t.Fatalf("topic listeners = %d, listens = %d, want 1, 1", serviceBus.topicListeners, serviceBus.topicListener.listens)
	}
	if got := logger.output.String(); got != "Worker working\n" {
		t.Fatalf("output = %q, want %q", got, "Worker working\n")
	}
}
