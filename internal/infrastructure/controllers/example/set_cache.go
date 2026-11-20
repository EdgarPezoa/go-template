package example

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgarpezoa/go-template/internal/application/types"
	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/request"
	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type SetCacheEndpoint struct {
	execute   func(context.Context, types.CacheInput) (types.CacheOutput, error)
	validator *validation.Validator
}

func (endpoint *SetCacheEndpoint) Handle(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	urlParameterKey := httpRequest.PathValue("key")
	input := request.SetCacheRequestDto{Key: urlParameterKey}
	if err := httptransport.Decode(responseWriter, httpRequest, &input); err != nil {
		return err
	}
	if err := endpoint.validateInput(input); err != nil {
		return err
	}
	output, err := endpoint.execute(httpRequest.Context(), types.CacheInput{Key: input.Key, Value: input.Value})
	if err != nil {
		return err
	}
	result := response.SetCacheResponseDto{Key: output.Key, Value: output.Value}
	if err := endpoint.validateOutput(result); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusCreated, result)
	return nil
}

func (endpoint *SetCacheEndpoint) validateOutput(output response.SetCacheResponseDto) error {
	return endpoint.validator.Validate(output)
}

func (endpoint *SetCacheEndpoint) validateInput(input request.SetCacheRequestDto) error {
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	return nil
}
