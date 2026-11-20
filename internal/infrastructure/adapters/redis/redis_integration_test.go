//go:build integration

package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRedisIntegration(t *testing.T) {
	address := os.Getenv("REDIS_INTEGRATION_ADDRESS")
	if address == "" {
		address = "localhost:6379"
	}
	redis, err := NewRedisAdapter(address, os.Getenv("REDIS_PASSWORD"), 0)
	if err != nil {
		t.Fatalf("connect to Redis at %s: %v", address, err)
	}
	defer redis.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := fmt.Sprintf("go-template-integration-%d", time.Now().UnixNano())
	value := fmt.Sprintf("integration-value-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = redis.exec(cleanupContext, "DEL", key)
	})
	normalizedKey, err := redis.Set(ctx, " "+key+" ", value)
	if err != nil {
		t.Fatal(err)
	}
	if normalizedKey != key {
		t.Fatalf("normalized key = %q, want %q", normalizedKey, key)
	}
	gotKey, got, err := redis.Get(ctx, " "+key+" ")
	if err != nil {
		t.Fatal(err)
	}
	if gotKey != key {
		t.Fatalf("get key = %q, want %q", gotKey, key)
	}
	if got != value {
		t.Fatalf("value = %q, want %q", got, value)
	}

	hashKey := fmt.Sprintf("go-template-integration-hash-%d", time.Now().UnixNano())
	hashField := fmt.Sprintf("integration-field-%d", time.Now().UnixNano())
	hashValue := fmt.Sprintf("integration-hash-value-%d", time.Now().UnixNano())
	normalizedHashKey := hashKey
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = redis.exec(cleanupContext, "DEL", normalizedHashKey)
	})
	normalizedHashKey, err = redis.HSet(ctx, " "+hashKey+" ", hashField, hashValue)
	if err != nil {
		t.Fatal(err)
	}
	if normalizedHashKey != hashKey {
		t.Fatalf("normalized hash key = %q, want %q", normalizedHashKey, hashKey)
	}
	gotHashKey, gotHashValue, err := redis.HGet(ctx, " "+hashKey+" ", hashField)
	if err != nil {
		t.Fatal(err)
	}
	if gotHashKey != hashKey {
		t.Fatalf("hash get key = %q, want %q", gotHashKey, hashKey)
	}
	if gotHashValue != hashValue {
		t.Fatalf("hash value = %q, want %q", gotHashValue, hashValue)
	}
}
