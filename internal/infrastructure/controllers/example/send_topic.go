package example

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/request"
	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type SendTopicEndpoint struct {
	execute   func(context.Context, string) error
	validator *validation.Validator
}

func (endpoint *SendTopicEndpoint) Handle(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var input request.SendTopicRequestDto
	if err := httptransport.Decode(responseWriter, httpRequest, &input); err != nil {
		return err
	}
	if err := endpoint.validateInput(input); err != nil {
		return err
	}
	if err := endpoint.execute(httpRequest.Context(), input.Message); err != nil {
		return err
	}
	output := response.SendTopicResponseDto{Accepted: true}
	if err := endpoint.validateOutput(output); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusAccepted, output)
	return nil
}

func (endpoint *SendTopicEndpoint) validateInput(input request.SendTopicRequestDto) error {
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	return nil
}

func (endpoint *SendTopicEndpoint) validateOutput(output response.SendTopicResponseDto) error {
	return endpoint.validator.Validate(output)
}
