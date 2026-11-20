package example

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/edgarpezoa/go-template/internal/application/types"
	exampleusecase "github.com/edgarpezoa/go-template/internal/application/use-cases/example"
	"github.com/edgarpezoa/go-template/internal/config"
	"github.com/edgarpezoa/go-template/internal/domain/models"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type redisPortFake struct{}

func (redisPortFake) Get(context.Context, string) (string, string, error)      { return "", "", nil }
func (redisPortFake) Set(context.Context, string, string) (string, error)      { return "", nil }
func (redisPortFake) HGet(context.Context, string, string) (string, string, error) { return "", "", nil }
func (redisPortFake) HSet(context.Context, string, string, string) (string, error) { return "", nil }
func (redisPortFake) Close() error                                             { return nil }

type messagePublisherPortFake struct{}

func (messagePublisherPortFake) Publish(context.Context, string) error { return nil }
func (messagePublisherPortFake) Close(context.Context) error           { return nil }

type serviceBusPortFake struct{}

type databasePortFake struct{}

func (databasePortFake) DeleteOne(context.Context, string, any) (ports.DeleteResult, error) {
	return ports.DeleteResult{}, nil
}
func (databasePortFake) DeleteMany(context.Context, string, any) (ports.DeleteResult, error) {
	return ports.DeleteResult{}, nil
}
func (databasePortFake) FindOne(_ context.Context, _ string, _ any, result any) error {
	*result.(*models.ExampleValue) = models.ExampleValue{ID: "507f1f77bcf86cd799439011", Value: "value"}
	return nil
}
func (databasePortFake) Find(context.Context, string, any, any) error { return nil }
func (databasePortFake) Insert(context.Context, string, any) (string, error) {
	return "507f1f77bcf86cd799439011", nil
}
func (databasePortFake) Upsert(context.Context, string, any, any) (ports.UpsertResult, error) {
	return ports.UpsertResult{}, nil
}

func (serviceBusPortFake) CreateQueueSender(string) (ports.MessagePublisherPort, error) {
	return messagePublisherPortFake{}, nil
}
func (serviceBusPortFake) CreateTopicSender(string) (ports.MessagePublisherPort, error) {
	return messagePublisherPortFake{}, nil
}
func (serviceBusPortFake) CreateQueueListener(string) (ports.MessageListenerPort, error) {
	return nil, nil
}
func (serviceBusPortFake) CreateSubscriptionListener(string, string) (ports.MessageListenerPort, error) {
	return nil, nil
}

func validServiceBusConfig() config.ServiceBusConfig {
	return config.ServiceBusConfig{Queue: "queue", Topic: "topic"}
}

func TestNewExampleControllerSucceedsWithValidConfiguration(t *testing.T) {
	controller, err := NewExampleController(redisPortFake{}, serviceBusPortFake{}, databasePortFake{}, validServiceBusConfig())
	if err != nil {
		t.Fatalf("NewExampleController() error = %v", err)
	}
	if controller == nil {
		t.Fatal("expected controller")
	}
	for name, endpoint := range map[string]interface{}{
		"get cache":        controller.endpoints.getCache,
		"set cache":        controller.endpoints.setCache,
		"get cache hash":   controller.endpoints.getCacheHash,
		"set cache hash":   controller.endpoints.setCacheHash,
		"send queue":       controller.endpoints.sendQueue,
		"send topic":       controller.endpoints.sendTopic,
		"probe validation": controller.endpoints.probeValidation,
		"insert database": controller.endpoints.insertDatabase,
		"find database": controller.endpoints.findDatabase,
	} {
		if endpoint == nil {
			t.Errorf("%s endpoint is nil", name)
		}
	}
	if controller.endpoints.getCache.execute == nil ||
		controller.endpoints.setCache.execute == nil ||
		controller.endpoints.getCacheHash.execute == nil ||
		controller.endpoints.setCacheHash.execute == nil ||
		controller.endpoints.sendQueue.execute == nil ||
		controller.endpoints.sendTopic.execute == nil ||
		controller.endpoints.probeValidation.execute == nil ||
		controller.endpoints.insertDatabase.execute == nil ||
		controller.endpoints.findDatabase.execute == nil {
		t.Fatal("expected every endpoint execute callback to be wired")
	}
}

func TestExampleControllerRoutesRegistersExplicitPatterns(t *testing.T) {
	controller, err := NewExampleController(redisPortFake{}, serviceBusPortFake{}, databasePortFake{}, validServiceBusConfig())
	if err != nil {
		t.Fatalf("NewExampleController() error = %v", err)
	}
	mux := http.NewServeMux()
	controller.Routes(mux)

	for _, test := range []struct {
		method  string
		path    string
		pattern string
	}{
		{method: http.MethodGet, path: "/cache/item", pattern: "GET /cache/{key}"},
		{method: http.MethodPut, path: "/cache/item", pattern: "PUT /cache/{key}"},
		{method: http.MethodGet, path: "/cache/hash/item", pattern: "GET /cache/hash/{key}"},
		{method: http.MethodPut, path: "/cache/hash/item", pattern: "PUT /cache/hash/{key}"},
		{method: http.MethodPost, path: "/service-bus/queue/messages", pattern: "POST /service-bus/queue/messages"},
		{method: http.MethodPost, path: "/service-bus/topic/messages", pattern: "POST /service-bus/topic/messages"},
		{method: http.MethodPost, path: "/validation/probe", pattern: "POST /validation/probe"},
		{method: http.MethodPost, path: "/database/examples", pattern: "POST /database/examples"},
		{method: http.MethodGet, path: "/database/examples/507f1f77bcf86cd799439011", pattern: "GET /database/examples/{id}"},
	} {
		_, pattern := mux.Handler(httptest.NewRequest(test.method, test.path, nil))
		if pattern != test.pattern {
			t.Errorf("route %q matched pattern %q, want %q", test.path, pattern, test.pattern)
		}
	}

	for _, test := range []struct {
		method  string
		pattern string
	}{
		{method: http.MethodGet, pattern: "GET /cache/{key}"},
		{method: http.MethodPut, pattern: "PUT /cache/{key}"},
	} {
		_, pattern := mux.Handler(httptest.NewRequest(test.method, "/cache/hash", nil))
		if pattern != test.pattern {
			t.Errorf("cache hash root matched pattern %q, want %q", pattern, test.pattern)
		}
	}
}

func TestDatabaseInsertEndpointReturnsCreatedContract(t *testing.T) {
	endpoint := &InsertDatabaseEndpoint{
		execute: func(_ context.Context, input models.ExampleValue) (models.ExampleValue, error) {
			return models.ExampleValue{ID: "id", Value: input.Value}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodPost, "/database/examples", strings.NewReader(`{"value":"value"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	var output struct { ID string `json:"id"`; Value string `json:"value"` }
	if err := json.Unmarshal(recorder.Body.Bytes(), &output); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if output.ID != "id" || output.Value != "value" {
		t.Fatalf("response = %+v", output)
	}
}

func TestDatabaseFindEndpointReturnsContractForPathID(t *testing.T) {
	endpoint := &FindDatabaseEndpoint{
		execute: func(_ context.Context, id string) (models.ExampleValue, error) {
			if id != "id" {
				t.Fatalf("id = %q, want id", id)
			}
			return models.ExampleValue{ID: id, Value: "value"}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodGet, "/database/examples/id", nil)
	request.SetPathValue("id", "id")
	recorder := httptest.NewRecorder()
	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var output struct { ID string `json:"id"`; Value string `json:"value"` }
	if err := json.Unmarshal(recorder.Body.Bytes(), &output); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if output.ID != "id" || output.Value != "value" {
		t.Fatalf("response = %+v", output)
	}
}

func TestDatabaseInsertEndpointRejectsMalformedAndInvalidRequestsBeforeExecution(t *testing.T) {
	for _, body := range []string{`{"value":`, `{"value":""}`} {
		executed := false
		endpoint := &InsertDatabaseEndpoint{
			execute: func(context.Context, models.ExampleValue) (models.ExampleValue, error) {
				executed = true
				return models.ExampleValue{}, nil
			},
			validator: validation.NewValidator(),
		}
		request := httptest.NewRequest(http.MethodPost, "/database/examples", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
		if recorder.Code != http.StatusBadRequest || executed {
			t.Fatalf("body %q: status = %d, executed = %v; want 400 and false", body, recorder.Code, executed)
		}
	}
}

func TestDatabaseInsertEndpointMapsUseCaseError(t *testing.T) {
	executed := false
	endpoint := &InsertDatabaseEndpoint{
			execute: func(context.Context, models.ExampleValue) (models.ExampleValue, error) {
			executed = true
			return models.ExampleValue{}, errors.New("database unavailable")
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodPost, "/database/examples", strings.NewReader(`{"value":"value"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError || !executed {
		t.Fatalf("status = %d, executed = %v; want 500 and true", recorder.Code, executed)
	}
}

func TestDatabaseFindEndpointRejectsInvalidPathDTOBeforeExecution(t *testing.T) {
	executed := false
	endpoint := &FindDatabaseEndpoint{
		execute: func(context.Context, string) (models.ExampleValue, error) {
			executed = true
			return models.ExampleValue{}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodGet, "/database/examples/", nil)
	request.SetPathValue("id", "")
	recorder := httptest.NewRecorder()
	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || executed {
		t.Fatalf("status = %d, executed = %v; want 400 and false", recorder.Code, executed)
	}
}

func TestDatabaseFindEndpointMapsUseCaseErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: ports.ErrNotFound, status: http.StatusNotFound},
		{name: "internal", err: errors.New("database unavailable"), status: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			executed := false
			endpoint := &FindDatabaseEndpoint{
				execute: func(context.Context, string) (models.ExampleValue, error) {
					executed = true
					return models.ExampleValue{}, test.err
				},
				validator: validation.NewValidator(),
			}
			request := httptest.NewRequest(http.MethodGet, "/database/examples/id", nil)
			request.SetPathValue("id", "id")
			recorder := httptest.NewRecorder()
			httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
			if recorder.Code != test.status || !executed {
				t.Fatalf("status = %d, executed = %v; want %d and true", recorder.Code, executed, test.status)
			}
		})
	}
}

func TestCacheEndpointRejectsBodyKeyAndUsesPathKey(t *testing.T) {
	executed := false
	endpoint := &SetCacheEndpoint{
		execute: func(_ context.Context, _ types.CacheInput) (types.CacheOutput, error) {
			executed = true
			return types.CacheOutput{Key: "orders", Value: "stored"}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodPut, "/cache/orders", strings.NewReader(`{"key":"other","value":"stored"}`))
	request.SetPathValue("key", "orders")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	err := endpoint.Handle(recorder, request)
	if err == nil {
		t.Fatal("expected a nil use-case panic boundary to be avoided by validation")
	}
	var invalidRequest httptransport.InvalidRequestError
	if !errors.As(err, &invalidRequest) {
		t.Fatalf("expected body key to be rejected, got %v", err)
	}
	if executed {
		t.Fatal("execute called for invalid body key")
	}
}

func TestCacheEndpointsExtractStringPathKey(t *testing.T) {
	endpoint := &GetCacheEndpoint{execute: func(_ context.Context, key string) (types.CacheOutput, error) {
		if key != "plain" {
			t.Fatalf("key = %q, want plain", key)
		}
		return types.CacheOutput{Key: "plain", Value: "value"}, nil
	}}
	request := httptest.NewRequest(http.MethodGet, "/cache/plain", nil)
	request.SetPathValue("key", "plain")
	err := endpoint.Handle(httptest.NewRecorder(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestCacheHashEndpointExtractsHashPathKey(t *testing.T) {
	endpoint := &GetCacheHashEndpoint{execute: func(_ context.Context, key string) (types.CacheOutput, error) {
		if key != "hashed" {
			t.Fatalf("key = %q, want hashed", key)
		}
		return types.CacheOutput{Key: "hashed", Value: "value"}, nil
	}}
	request := httptest.NewRequest(http.MethodGet, "/cache/hash/hashed", nil)
	request.SetPathValue("key", "hashed")
	err := endpoint.Handle(httptest.NewRecorder(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
}

func TestCacheEndpointRejectsMissingPathKeyBeforeExecution(t *testing.T) {
	executed := false
	endpoint := &GetCacheEndpoint{
		execute: func(_ context.Context, _ string) (types.CacheOutput, error) {
			executed = true
			return types.CacheOutput{}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodGet, "/cache/", nil)
	request.SetPathValue("key", "")

	err := endpoint.Handle(httptest.NewRecorder(), request)
	if err == nil || !errors.Is(err, httptransport.ErrValidation) {
		t.Fatalf("error = %v, want validation error", err)
	}
	if executed {
		t.Fatal("execute called for invalid path key")
	}
}

func TestCacheEndpointResponseContractFailureMapsToInternalError(t *testing.T) {
	endpoint := &GetCacheEndpoint{
		execute: func(_ context.Context, _ string) (types.CacheOutput, error) {
			return types.CacheOutput{}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodGet, "/cache/item", nil)
	request.SetPathValue("key", "item")
	recorder := httptest.NewRecorder()

	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}

func TestExampleRoutesUseNativeMethodHandling(t *testing.T) {
	controller, err := NewExampleController(redisPortFake{}, serviceBusPortFake{}, databasePortFake{}, validServiceBusConfig())
	if err != nil { t.Fatalf("NewExampleController() error = %v", err) }
	mux := http.NewServeMux()
	controller.Routes(mux)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPatch, "/cache/item", nil))
	if recorder.Code != http.StatusMethodNotAllowed { t.Fatalf("status = %d, want 405", recorder.Code) }
	if recorder.Header().Get("Allow") != "GET, HEAD, PUT" { t.Fatalf("Allow = %q, want GET, HEAD, PUT", recorder.Header().Get("Allow")) }
}

func TestExampleRoutesUseNativeNotFoundHandling(t *testing.T) {
	controller, err := NewExampleController(redisPortFake{}, serviceBusPortFake{}, databasePortFake{}, validServiceBusConfig())
	if err != nil { t.Fatalf("NewExampleController() error = %v", err) }
	mux := http.NewServeMux()
	controller.Routes(mux)
	for _, path := range []string{"/missing", "/cache/"} {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusNotFound { t.Fatalf("%s = %d, want 404", path, recorder.Code) }
	}
}

func TestProbeValidationEndpointReturnsDistinctValidatedResponse(t *testing.T) {
	endpoint := &ProbeValidationEndpoint{
		execute: func(_ context.Context, input types.ValidationProbeInput) (types.ValidationProbeOutput, error) {
			return exampleusecase.NewProbeValidationUseCase().Execute(context.Background(), input)
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodPost, "/validation/probe", bytes.NewBufferString(`{"name":"case","enabled":true,"attempts":3,"mode":"strict","profile":{"name":"Ada"},"labels":[]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var output map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &output); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if output["result"] != "accepted" || output["summary"] != "case/strict" {
		t.Fatalf("response = %#v", output)
	}
	if labels, ok := output["labelCodes"].([]any); !ok || len(labels) != 0 {
		t.Fatalf("labelCodes = %#v, want []", output["labelCodes"])
	}
}

func TestProbeValidationEndpointRejectsInvalidActiveFieldsBeforeExecution(t *testing.T) {
	executed := false
	endpoint := &ProbeValidationEndpoint{
		execute: func(context.Context, types.ValidationProbeInput) (types.ValidationProbeOutput, error) {
			executed = true
			return types.ValidationProbeOutput{}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodPost, "/validation/probe", bytes.NewBufferString(`{"name":"","enabled":"yes","attempts":-1,"mode":"other","profile":{"name":""},"labels":[{"code":""}]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || executed {
		t.Fatalf("status = %d, executed = %v; want 400 and false", recorder.Code, executed)
	}
}

func TestProbeValidationEndpointAcceptsDisabledWrongShapedFields(t *testing.T) {
	executed := false
	endpoint := &ProbeValidationEndpoint{
		execute: func(_ context.Context, input types.ValidationProbeInput) (types.ValidationProbeOutput, error) {
			executed = true
			return types.ValidationProbeOutput{
				Outcome: "review", Summary: "case/basic", Subject: types.ValidationProbeSubject{DisplayName: "Ada"},
				LabelCodes: []string{}, Processed: true,
			}, nil
		},
		validator: validation.NewValidator(),
	}
	body := `{"name":"case","disabledRequired":123,"enabled":false,"disabledString":123,"attempts":1,"disabledBoolean":"wrong","mode":"basic","disabledNumber":"wrong","profile":{"name":"Ada"},"disabledStruct":"wrong","labels":[],"disabledArray":"wrong"}`
	request := httptest.NewRequest(http.MethodPost, "/validation/probe", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !executed {
		t.Fatalf("status = %d, executed = %v; want 200 and true", recorder.Code, executed)
	}
}

func TestProbeValidationEndpointMapsInvalidOutputToInternalError(t *testing.T) {
	endpoint := &ProbeValidationEndpoint{
		execute: func(context.Context, types.ValidationProbeInput) (types.ValidationProbeOutput, error) {
			return types.ValidationProbeOutput{Outcome: "invalid"}, nil
		},
		validator: validation.NewValidator(),
	}
	request := httptest.NewRequest(http.MethodPost, "/validation/probe", bytes.NewBufferString(`{"name":"case","enabled":true,"attempts":1,"mode":"basic","profile":{"name":"Ada"},"labels":[]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	httptransport.Handler(endpoint.Handle).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
}

func TestMapErrorCoversControllerCategories(t *testing.T) {
	recorder := httptest.NewRecorder()
	httptransport.Handler(func(http.ResponseWriter, *http.Request) error {
		return httptransport.ErrValidation
	}).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}
