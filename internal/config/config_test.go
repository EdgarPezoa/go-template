package config

import (
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{
		"API_PORT",
		"REDIS_ADDRESS",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"AZURE_SERVICEBUS_CONNECTION_STRING",
		"AZURE_SERVICEBUS_QUEUE",
		"AZURE_SERVICEBUS_TOPIC",
		"AZURE_SERVICEBUS_SUBSCRIPTION",
		"MONGODB_URL",
		"MONGODB_USER",
		"MONGODB_PASSWORD",
		"MONGODB_ARGS",
		"MONGODB_DATABASE",
		"MONGODB_PORT",
		"MONGODB_SERVER_SELECTION_TIMEOUT_SECONDS",
		"MONGODB_PING_TIMEOUT_SECONDS",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("REDIS_ADDRESS", "localhost:6379")
	t.Setenv("AZURE_SERVICEBUS_CONNECTION_STRING", "Endpoint=sb://example/")

	config, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if config.API.Port != 3000 || config.ServiceBus.Queue != "go-template-messages" {
		t.Fatalf("defaults not applied: %+v", config)
	}
	if config.Mongo.URL != "localhost" || config.Mongo.Port != 27017 || config.Mongo.Database != "go-template" || config.Mongo.Args != "" || config.Mongo.ServerSelectionTimeoutSeconds != 3 || config.Mongo.PingTimeoutSeconds != 3 {
		t.Fatalf("mongodb defaults not applied: %+v", config.Mongo)
	}
}

func TestLoadRejectsMongoHostAndUnpairedCredentials(t *testing.T) {
	for _, host := range []string{"mongodb://localhost", "localhost:27017", "localhost/path", "localhost?query=1"} {
		t.Run(host, func(t *testing.T) {
			t.Setenv("MONGODB_URL", host)
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "MONGODB_URL") {
				t.Fatalf("expected host-only validation error, got %v", err)
			}
		})
	}
	t.Setenv("MONGODB_URL", "localhost")
	t.Setenv("MONGODB_USER", "user")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "MONGODB_USER") {
		t.Fatalf("expected paired credential validation error, got %v", err)
	}
}

func TestLoadCanonicalizesMongoArgs(t *testing.T) {
	t.Setenv("MONGODB_ARGS", "z=last&a=hello%20world&a=first")
	config, err := Load()
	if err != nil || config.Mongo.Args != "a=hello+world&a=first&z=last" {
		t.Fatalf("args = %q, err = %v", config.Mongo.Args, err)
	}
}

func TestLoadRejectsInvalidMongoArgs(t *testing.T) {
	for _, args := range []string{"bad=%zz", "a=1;b=2"} {
		t.Run(args, func(t *testing.T) {
			t.Setenv("MONGODB_ARGS", args)
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "MONGODB_ARGS") {
				t.Fatalf("expected args validation error, got %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidInteger(t *testing.T) {
	t.Setenv("API_PORT", "not-a-number")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "API_PORT") {
		t.Fatalf("expected annotated API_PORT parse error, got %v", err)
	}
}

func TestLoadRejectsAPIPortOutsideValidRange(t *testing.T) {
	for _, port := range []string{"0", "65536"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("API_PORT", port)
			if _, err := Load(); err == nil || !strings.Contains(err.Error(), "API.Port") {
				t.Fatalf("expected API port validation error, got %v", err)
			}
		})
	}
}

func TestLoadRejectsInvalidMongoTimeouts(t *testing.T) {
	for _, name := range []string{"MONGODB_SERVER_SELECTION_TIMEOUT_SECONDS", "MONGODB_PING_TIMEOUT_SECONDS"} {
		for _, value := range []string{"not-a-number", "0", "-1"} {
			t.Run(name+"="+value, func(t *testing.T) {
				t.Setenv(name, value)
				if _, err := Load(); err == nil || !strings.Contains(err.Error(), name) {
					t.Fatalf("expected timeout validation error, got %v", err)
				}
			})
		}
	}
}
