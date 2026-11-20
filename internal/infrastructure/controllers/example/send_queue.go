package example

import (
	"context"
	"net/http"

	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/request"
	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type SendQueueEndpoint struct {
	execute   func(context.Context, string) error
	validator *validation.Validator
}

func (endpoint *SendQueueEndpoint) Handle(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var input request.SendQueueRequestDto
	if err := httptransport.Decode(responseWriter, httpRequest, &input); err != nil {
		return err
	}
	if err := endpoint.validateInput(input); err != nil {
		return err
	}
	if err := endpoint.execute(httpRequest.Context(), input.Message); err != nil {
		return err
	}
	output := response.SendQueueResponseDto{Accepted: true}
	if err := endpoint.validateOutput(output); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusAccepted, output)
	return nil
}

func (endpoint *SendQueueEndpoint) validateInput(input request.SendQueueRequestDto) error {
	return endpoint.validator.Validate(input);
}

func (endpoint *SendQueueEndpoint) validateOutput(output response.SendQueueResponseDto) error {
	return endpoint.validator.Validate(output)
}
