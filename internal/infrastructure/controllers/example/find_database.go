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

type FindDatabaseEndpoint struct {
	execute   func(context.Context, string) (models.ExampleValue, error)
	validator *validation.Validator
}

func (endpoint *FindDatabaseEndpoint) Handle(w http.ResponseWriter, r *http.Request) error {
	input := request.FindDatabaseRequestDto{ID: r.PathValue("id")}
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	output, err := endpoint.execute(r.Context(), input.ID)
	if err != nil {
		return err
	}
	result := response.FindDatabaseResponseDto{ID: output.ID, Value: output.Value}
	if err := endpoint.validator.Validate(result); err != nil {
		return err
	}
	httptransport.WriteJSON(w, http.StatusOK, result)
	return nil
}
