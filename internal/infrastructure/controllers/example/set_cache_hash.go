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

type SetCacheHashEndpoint struct {
	execute   func(context.Context, types.CacheInput) (types.CacheOutput, error)
	validator *validation.Validator
}

func (endpoint *SetCacheHashEndpoint) Handle(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	urlParameterKey := httpRequest.PathValue("key")
	input := request.SetCacheHashRequestDto{Key: urlParameterKey}
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
	result := response.SetCacheHashResponseDto{Key: output.Key, Value: output.Value}
	if err := endpoint.validateOutput(result); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusCreated, result)
	return nil
}

func (endpoint *SetCacheHashEndpoint) validateOutput(output response.SetCacheHashResponseDto) error {
	return endpoint.validator.Validate(output)
}

func (endpoint *SetCacheHashEndpoint) validateInput(input request.SetCacheHashRequestDto) error {
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	return nil
}
