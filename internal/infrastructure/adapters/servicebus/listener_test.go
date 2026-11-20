package servicebus

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeSender struct {
	close func() error
}

func (sender *fakeSender) Send(context.Context, string) error {
	return nil
}

func (sender *fakeSender) Close(context.Context) error {
	return sender.close()
}

type fakeReceiver struct {
	receive  func() ([]receivedMessage, error)
	complete func() error
	abandon  func() error
	close    func() error
}

func (receiver *fakeReceiver) Receive(context.Context) ([]receivedMessage, error) {
	return receiver.receive()
}

func (receiver *fakeReceiver) Complete(context.Context, receivedMessage) error {
	return receiver.complete()
}

func (receiver *fakeReceiver) Abandon(context.Context, receivedMessage) error {
	return receiver.abandon()
}

func (receiver *fakeReceiver) Close(context.Context) error {
	return receiver.close()
}

func TestListenerRequiresConnection(t *testing.T) {
	if _, err := NewServiceBusAdapter(""); err == nil {
		t.Fatal("expected connection error")
	}
}
func TestQueueListenerRequiresQueue(t *testing.T) {
	factory, err := NewServiceBusAdapter("Endpoint=sb://localhost;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;")
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close(context.Background())
	if _, err := factory.CreateQueueListener(""); err == nil {
		t.Fatal("expected queue error")
	}
}
func TestTopicListenerRequiresTopicAndSubscription(t *testing.T) {
	factory, err := NewServiceBusAdapter("Endpoint=sb://localhost;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;")
	if err != nil {
		t.Fatal(err)
	}
	defer factory.Close(context.Background())
	if _, err := factory.CreateSubscriptionListener("", "s"); err == nil {
		t.Fatal("expected topic error")
	}
	if _, err := factory.CreateSubscriptionListener("t", ""); err == nil {
		t.Fatal("expected subscription error")
	}
}

func TestListenerProcessErrorAbandonsAndContinues(t *testing.T) {
	processErr := errors.New("process")
	receiveErr := errors.New("stop")
	processed := []string{}
	abandoned := 0
	errorsSeen := []error{}
	receiveCount := 0
	listener := &listener{receiver: &fakeReceiver{
		receive: func() ([]receivedMessage, error) {
			receiveCount++
			if receiveCount == 1 {
				return []receivedMessage{{body: "first"}, {body: "second"}}, nil
			}
			return nil, receiveErr
		},
		complete: func() error { return nil },
		abandon: func() error {
			abandoned++
			return nil
		},
		close: func() error { return nil },
	}}

	err := listener.Listen(context.Background(), func(_ context.Context, message string) error {
		processed = append(processed, message)
		if message == "first" {
			return processErr
		}
		return nil
	}, func(_ context.Context, err error) {
		errorsSeen = append(errorsSeen, err)
	})

	if !errors.Is(err, receiveErr) || abandoned != 1 ||
		!reflect.DeepEqual(processed, []string{"first", "second"}) ||
		!reflect.DeepEqual(errorsSeen, []error{processErr, receiveErr}) {
		t.Fatalf("err=%v abandoned=%d processed=%v errors=%v", err, abandoned, processed, errorsSeen)
	}
}

func TestListenerReceiveErrorReportsAndReturns(t *testing.T) {
	receiveErr := errors.New("receive")
	errorsSeen := []error{}
	listener := &listener{receiver: &fakeReceiver{
		receive: func() ([]receivedMessage, error) { return nil, receiveErr },
		complete: func() error { return nil },
		abandon: func() error { return nil },
		close: func() error { return nil },
	}}

	err := listener.Listen(context.Background(), nil, func(_ context.Context, err error) {
		errorsSeen = append(errorsSeen, err)
	})
	if !errors.Is(err, receiveErr) || !reflect.DeepEqual(errorsSeen, []error{receiveErr}) {
		t.Fatalf("err=%v errors=%v", err, errorsSeen)
	}
}

func TestListenerCompleteErrorReportsAndReturns(t *testing.T) {
	completeErr := errors.New("complete")
	errorsSeen := []error{}
	listener := &listener{receiver: &fakeReceiver{
		receive: func() ([]receivedMessage, error) { return []receivedMessage{{body: "message"}}, nil },
		complete: func() error { return completeErr },
		abandon: func() error { return nil },
		close: func() error { return nil },
	}}

	err := listener.Listen(context.Background(), func(context.Context, string) error {
		return nil
	}, func(_ context.Context, err error) {
		errorsSeen = append(errorsSeen, err)
	})
	if !errors.Is(err, completeErr) || !reflect.DeepEqual(errorsSeen, []error{completeErr}) {
		t.Fatalf("err=%v errors=%v", err, errorsSeen)
	}
}

func TestListenerAbandonFailureReportsProcessAndAbandonErrors(t *testing.T) {
	processErr := errors.New("process")
	abandonErr := errors.New("abandon")
	errorsSeen := []error{}
	listener := &listener{receiver: &fakeReceiver{
		receive: func() ([]receivedMessage, error) { return []receivedMessage{{body: "message"}}, nil },
		complete: func() error { return nil },
		abandon: func() error { return abandonErr },
		close: func() error { return nil },
	}}

	err := listener.Listen(context.Background(), func(context.Context, string) error {
		return processErr
	}, func(_ context.Context, err error) {
		errorsSeen = append(errorsSeen, err)
	})
	if !errors.Is(err, abandonErr) ||
		!reflect.DeepEqual(errorsSeen, []error{processErr, abandonErr}) {
		t.Fatalf("err=%v errors=%v", err, errorsSeen)
	}
}
