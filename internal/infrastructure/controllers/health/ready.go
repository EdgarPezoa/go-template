package health

import (
	"net/http"

	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/health/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type readyEndpoint struct {
	readiness func() bool
	validator  *validation.Validator
}

func (endpoint *readyEndpoint) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	status := response.ReadyResponseDto{Status: "ready"}
	httpStatus := http.StatusOK
	if !endpoint.readiness() {
		status.Status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}
	if err := endpoint.validateOutput(status); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, httpStatus, status)
	return nil
}

func (endpoint *readyEndpoint) validateOutput(output response.ReadyResponseDto) error {
	return endpoint.validator.Validate(output)
}
