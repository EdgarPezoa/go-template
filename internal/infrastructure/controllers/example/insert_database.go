package example

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edgarpezoa/go-template/internal/domain/models"
	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/request"
	"github.com/edgarpezoa/go-template/internal/infrastructure/controllers/example/dto/response"
	"github.com/edgarpezoa/go-template/internal/shared/controller/httptransport"
	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type InsertDatabaseEndpoint struct {
	execute   func(context.Context, models.ExampleValue) (models.ExampleValue, error)
	validator *validation.Validator
}

func (endpoint *InsertDatabaseEndpoint) Handle(w http.ResponseWriter, r *http.Request) error {
	var input request.InsertDatabaseRequestDto
	if err := httptransport.Decode(w, r, &input); err != nil {
		return err
	}
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	output, err := endpoint.execute(r.Context(), models.ExampleValue{Value: input.Value})
	if err != nil {
		return err
	}
	result := response.InsertDatabaseResponseDto{ID: output.ID, Value: output.Value}
	if err := endpoint.validator.Validate(result); err != nil {
		return err
	}
	httptransport.WriteJSON(w, http.StatusCreated, result)
	return nil
}
