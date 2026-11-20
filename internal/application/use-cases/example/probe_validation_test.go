package example

import (
	"context"
	"testing"

	"github.com/edgarpezoa/go-template/internal/application/types"
)

func TestProbeValidationUseCaseMapsAcceptedInput(t *testing.T) {
	output, err := NewProbeValidationUseCase().Execute(context.Background(), types.ValidationProbeInput{
		Enabled:  true,
		Name:     "case",
		Attempts: 3,
		Mode:     "strict",
		Profile:  types.ValidationProbeProfile{Name: "Ada"},
		Labels:   []types.ValidationProbeLabel{{Code: "A"}, {Code: "B"}},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Outcome != "accepted" || output.Summary != "case/strict" || output.RemainingAttempts != 2 || !output.Processed {
		t.Fatalf("output = %#v", output)
	}
	if output.Subject.DisplayName != "Ada" {
		t.Fatalf("subject = %#v", output.Subject)
	}
	if len(output.LabelCodes) != 2 || output.LabelCodes[0] != "A" || output.LabelCodes[1] != "B" {
		t.Fatalf("label codes = %#v", output.LabelCodes)
	}
}

func TestProbeValidationUseCaseMapsReviewAndEmptyLabels(t *testing.T) {
	output, err := NewProbeValidationUseCase().Execute(context.Background(), types.ValidationProbeInput{
		Name:     "case",
		Attempts: 0,
		Mode:     "basic",
		Profile:  types.ValidationProbeProfile{Name: "Ada"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Outcome != "review" || output.RemainingAttempts != 0 || output.LabelCodes == nil || len(output.LabelCodes) != 0 {
		t.Fatalf("output = %#v", output)
	}
}
