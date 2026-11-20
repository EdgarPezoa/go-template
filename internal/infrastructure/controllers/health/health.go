package health

import (
	"net/http"

	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/health/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type healthEndpoint struct {
	validator *validation.Validator
}

func (endpoint *healthEndpoint) Handle(responseWriter http.ResponseWriter, request *http.Request) error {
	output := response.HealthResponseDto{Status: "ok"}
	if err := endpoint.validateOutput(output); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusOK, output)
	return nil
}

func (endpoint *healthEndpoint) validateOutput(output response.HealthResponseDto) error {
	return endpoint.validator.Validate(output)
}
