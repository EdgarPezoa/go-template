package servicebus

import "testing"

func TestPublisherRequiresConnection(t *testing.T) {
	if _, err := NewServiceBusAdapter(""); err == nil {
		t.Fatal("expected connection error")
	}
}
