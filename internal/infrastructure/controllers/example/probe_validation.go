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

type ProbeValidationEndpoint struct {
	execute   func(context.Context, types.ValidationProbeInput) (types.ValidationProbeOutput, error)
	validator *validation.Validator
}

func (endpoint *ProbeValidationEndpoint) Handle(responseWriter http.ResponseWriter, httpRequest *http.Request) error {
	var input request.ProbeValidationRequestDto
	if err := httptransport.Decode(responseWriter, httpRequest, &input); err != nil {
		return err
	}
	if err := endpoint.validateInput(input); err != nil {
		return err
	}
	output, err := endpoint.execute(httpRequest.Context(), types.ValidationProbeInput{
		Enabled:  input.Enabled,
		Name:     input.Name,
		Attempts: input.Attempts,
		Mode:     input.Mode,
		Profile:  types.ValidationProbeProfile{Name: input.Profile.Name},
		Labels:   probeValidationLabels(input.Labels),
	})
	if err != nil {
		return err
	}
	result := response.ProbeValidationResponseDto{
		Result:            output.Outcome,
		Summary:           output.Summary,
		RemainingAttempts: output.RemainingAttempts,
		Processed:         output.Processed,
		Subject:           response.ProbeValidationSubjectDto{DisplayName: output.Subject.DisplayName},
		LabelCodes:        output.LabelCodes,
	}
	if err := endpoint.validator.Validate(result); err != nil {
		return err
	}
	httptransport.WriteJSON(responseWriter, http.StatusOK, result)
	return nil
}

func (endpoint *ProbeValidationEndpoint) validateInput(input request.ProbeValidationRequestDto) error {
	if err := endpoint.validator.Validate(input); err != nil {
		return fmt.Errorf("%w: %v", httptransport.ErrValidation, err)
	}
	return nil
}

func probeValidationLabels(labels []request.ProbeValidationLabelDto) []types.ValidationProbeLabel {
	result := make([]types.ValidationProbeLabel, 0, len(labels))
	for _, label := range labels {
		result = append(result, types.ValidationProbeLabel{Code: label.Code})
	}
	return result
}
