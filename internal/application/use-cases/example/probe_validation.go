package example

import (
	"context"
	"fmt"

	"github.com/edgarpezoa/go-template/internal/application/types"
)

type ProbeValidationUseCase struct{}

func NewProbeValidationUseCase() *ProbeValidationUseCase {
	return &ProbeValidationUseCase{}
}

func (useCase *ProbeValidationUseCase) Execute(_ context.Context, input types.ValidationProbeInput) (types.ValidationProbeOutput, error) {
	outcome := "review"
	if input.Enabled {
		outcome = "accepted"
	}

	labelCodes := make([]string, 0, len(input.Labels))
	for _, label := range input.Labels {
		labelCodes = append(labelCodes, label.Code)
	}

	remainingAttempts := 0
	if input.Attempts > 0 {
		remainingAttempts = input.Attempts - 1
	}

	return types.ValidationProbeOutput{
		Outcome:           outcome,
		Summary:           fmt.Sprintf("%s/%s", input.Name, input.Mode),
		RemainingAttempts: remainingAttempts,
		Processed:         true,
		Subject:           types.ValidationProbeSubject{DisplayName: input.Profile.Name},
		LabelCodes:        labelCodes,
	}, nil
}
