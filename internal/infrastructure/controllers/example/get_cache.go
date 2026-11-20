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

type GetCacheEndpoint struct {
	execute   func(context.Context, string) (types.CacheOutput, error)
	validator *validation.Validator
}

func (endpoint *GetCacheEndpoint) Handle(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	input := request.GetCacheRequestDto{Key: httpRequest.PathValue("key")}
	if err := endpoint.validateInput(input); err != nil {
		return err
	}
	output, err := endpoint.execute(httpRequest.Context(), input.Key)
	if err != nil {
		return err
	}
	result := response.GetCacheResponseDto{Key: output.Key, Value: output.Value}
	if err := endpoint.validateOutput(result); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusOK, result)
	return nil
}

func (endpoint *GetCacheEndpoint) validateInput(input request.GetCacheRequestDto) error {
	if endpoint.validator == nil {
		return nil
	}
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	return nil
}

func (endpoint *GetCacheEndpoint) validateOutput(output response.GetCacheResponseDto) error {
	if endpoint.validator == nil {
		return nil
	}
	return endpoint.validator.Validate(output)
}
