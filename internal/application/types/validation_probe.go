package types

type ValidationProbeInput struct {
	Enabled  bool
	Name     string
	Attempts int
	Mode     string
	Profile  ValidationProbeProfile
	Labels   []ValidationProbeLabel
}

type ValidationProbeProfile struct {
	Name string
}

type ValidationProbeLabel struct {
	Code string
}

type ValidationProbeOutput struct {
	Outcome           string
	Summary           string
	RemainingAttempts int
	Processed         bool
	Subject           ValidationProbeSubject
	LabelCodes        []string
}

type ValidationProbeSubject struct {
	DisplayName string
}
