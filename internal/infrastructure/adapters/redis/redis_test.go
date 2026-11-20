package redis

import (
	"bufio"
	"errors"
	"strings"
	"testing"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

func TestRedisRequiresAddress(t *testing.T) {
	if _, err := NewRedisAdapter("", "", 0); err == nil {
		t.Fatal("expected address error")
	}
}

func TestNormalizeKey(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{input: " key ", want: "key"},
		{input: "key", want: "key"},
	} {
		got, err := normalizeKey(test.input)
		if err != nil || got != test.want {
			t.Fatalf("normalizeKey(%q) = %q, %v; want %q, nil", test.input, got, err, test.want)
		}
	}
	for _, key := range []string{"", " \t\n "} {
		if _, err := normalizeKey(key); !errors.Is(err, ports.ErrInvalidKey) {
			t.Fatalf("normalizeKey(%q) error = %v, want ErrInvalidKey", key, err)
		}
	}
}

func TestRedisOperationsRejectInvalidKeysBeforeExecution(t *testing.T) {
	redis := &RedisAdapter{}
	if _, _, err := redis.Get(nil, " \t"); !errors.Is(err, ports.ErrInvalidKey) {
		t.Fatalf("Get() error = %v, want ErrInvalidKey", err)
	}
	if _, err := redis.Set(nil, " \t", "value"); !errors.Is(err, ports.ErrInvalidKey) {
		t.Fatalf("Set() error = %v, want ErrInvalidKey", err)
	}
	if _, _, err := redis.HGet(nil, " \t", "value"); !errors.Is(err, ports.ErrInvalidKey) {
		t.Fatalf("HGet() error = %v, want ErrInvalidKey", err)
	}
	if _, err := redis.HSet(nil, " \t", "value", "value"); !errors.Is(err, ports.ErrInvalidKey) {
		t.Fatalf("HSet() error = %v, want ErrInvalidKey", err)
	}
}

func TestReadRESPRejectsMalformedBulkLength(t *testing.T) {
	_, err := readRESP(bufio.NewReader(strings.NewReader("$not-a-number\r\n")))
	if err == nil {
		t.Fatal("expected malformed bulk length error")
	}
}

func TestReadRESPReadsExactBulkPayload(t *testing.T) {
	value, err := readRESP(bufio.NewReader(strings.NewReader("$5\r\nhello\r\n")))
	if err != nil {
		t.Fatal(err)
	}
	if value != "hello" {
		t.Fatalf("value = %v, want hello", value)
	}
}
