package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/health/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

func TestHealthController(t *testing.T) {
	if NewHealthController() == nil {
		t.Fatal("nil controller")
	}
}

func TestHealthControllerInjectsValidator(t *testing.T) {
	controller := NewHealthController()
	if controller.validator == nil {
		t.Fatal("expected validator")
	}
	if (&healthEndpoint{validator: validation.NewValidator()}).validateOutput(response.HealthResponseDto{}) == nil {
		t.Fatal("expected invalid response contract")
	}
}

func TestHealthRoutesUseRenamedEndpoints(t *testing.T) {
	controller := NewHealthController(func() bool { return false })
	mux := http.NewServeMux()
	controller.Routes(mux)

	healthRecorder := httptest.NewRecorder()
	mux.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if healthRecorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthRecorder.Code, http.StatusOK)
	}

	readyRecorder := httptest.NewRecorder()
	mux.ServeHTTP(readyRecorder, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if readyRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("ready status = %d, want %d", readyRecorder.Code, http.StatusServiceUnavailable)
	}
}

func TestHealthRoutesRejectUnsupportedMethods(t *testing.T) {
	controller := NewHealthController()
	mux := http.NewServeMux()
	controller.Routes(mux)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/health", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	if recorder.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("Allow = %q, want GET, HEAD", recorder.Header().Get("Allow"))
	}
}

func TestHealthRoutesUseNativeGetHeadMatching(t *testing.T) {
	controller := NewHealthController()
	mux := http.NewServeMux()
	controller.Routes(mux)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodHead, "/health", nil))
	if recorder.Code != http.StatusOK { t.Fatalf("HEAD status = %d, want 200", recorder.Code) }
}
