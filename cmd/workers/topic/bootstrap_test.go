package main

import (
	"testing"

	"github.com/edgarpezoa/go-template/internal/config"
)

func TestNewWithoutServiceBusReturnsError(t *testing.T) {
	bootstrap, err := NewBootstrap(config.Config{})
	if err == nil {
		t.Fatal("expected service bus initialization error")
	}
	if bootstrap != nil {
		t.Fatal("bootstrap != nil")
	}
}
