package response

type ProbeValidationResponseDto struct {
	Result            string                    `json:"result" required:"true" enum:"accepted|review"`
	Summary           string                    `json:"summary" required:"true" string:"true"`
	RemainingAttempts int                       `json:"remainingAttempts" number:"true" min:"0"`
	Processed         bool                      `json:"processed" boolean:"true"`
	Subject           ProbeValidationSubjectDto `json:"subject" required:"true" struct:"true"`
	LabelCodes        []string                  `json:"labelCodes" required:"true" array:"true"`
}

type ProbeValidationSubjectDto struct {
	DisplayName string `json:"displayName" required:"true" string:"true"`
}
