package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCloseResourcesClosesAllInReverseOrder(t *testing.T) {
	var closed []string
	closers := []func(context.Context) error{
		func(context.Context) error {
			closed = append(closed, "redis")
			return nil
		},
		func(context.Context) error {
			closed = append(closed, "queue")
			return nil
		},
		func(context.Context) error {
			closed = append(closed, "topic")
			return nil
		},
		func(context.Context) error {
			closed = append(closed, "mongodb")
			return nil
		},
	}

	if err := closeResources(closers, context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := []string{"mongodb", "topic", "queue", "redis"}; !reflect.DeepEqual(closed, want) {
		t.Fatalf("closed = %v, want %v", closed, want)
	}
}

func TestServeMuxUsesNativeMissAndMethodHandling(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(http.ResponseWriter, *http.Request) {})

	missing := httptest.NewRecorder()
	mux.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/missing", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want %d", missing.Code, http.StatusNotFound)
	}

	unsupported := httptest.NewRecorder()
	mux.ServeHTTP(unsupported, httptest.NewRequest(http.MethodPost, "/health", nil))
	if unsupported.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unsupported status = %d, want %d", unsupported.Code, http.StatusMethodNotAllowed)
	}
	if unsupported.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", unsupported.Header().Get("Allow"))
	}
}
