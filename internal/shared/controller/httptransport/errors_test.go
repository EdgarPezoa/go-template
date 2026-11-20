package httptransport

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

func TestEvaluate(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", nil)
	invalidRequest := Decode(httptest.NewRecorder(), request, &struct{}{})
	tests := []struct {
		name     string
		err      error
		expected ErrorResponse
	}{
		{"invalid request", invalidRequest, ErrorResponse{Status: http.StatusBadRequest, Code: "invalid_request", Message: "invalid request body"}},
		{"port not found", ports.ErrNotFound, ErrorResponse{Status: http.StatusNotFound, Code: "not_found", Message: "value not found"}},
		{"invalid key", ports.ErrInvalidKey, ErrorResponse{Status: http.StatusBadRequest, Code: "validation_error", Message: "validation failed"}},
		{"wrapped validation", errors.Join(errors.New("context"), ErrValidation), ErrorResponse{Status: http.StatusBadRequest, Code: "validation_error", Message: "validation failed"}},
		{"unknown", errors.New("unknown"), ErrorResponse{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal server error"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if response := evaluate(test.err); response != test.expected {
				t.Fatalf("Evaluate() = %+v, want %+v", response, test.expected)
			}
		})
	}
}
