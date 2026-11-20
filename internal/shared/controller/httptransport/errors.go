package httptransport

import (
	"errors"
	"net/http"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

var (
	ErrValidation = errors.New("validation failed")
)

func evaluate(err error) ErrorResponse {
	var invalid InvalidRequestError
	if errors.As(err, &invalid) {
		return ErrorResponse{Status: http.StatusBadRequest, Code: "invalid_request", Message: "invalid request body"}
	}
	if errors.Is(err, ports.ErrNotFound) {
		return ErrorResponse{Status: http.StatusNotFound, Code: "not_found", Message: "value not found"}
	}
	if errors.Is(err, ports.ErrInvalidKey) {
		return ErrorResponse{Status: http.StatusBadRequest, Code: "validation_error", Message: "validation failed"}
	}
	if errors.Is(err, ports.ErrInvalidID) {
		return ErrorResponse{Status: http.StatusBadRequest, Code: "validation_error", Message: "validation failed"}
	}
	if errors.Is(err, ErrValidation) {
		return ErrorResponse{Status: http.StatusBadRequest, Code: "validation_error", Message: "validation failed"}
	}
	return ErrorResponse{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal server error"}
}
