package request

type ProbeValidationRequestDto struct {
	Name             string                    `json:"name" required:"true" string:"true"`
	DisabledRequired int                       `json:"disabledRequired" required:"false"`
	Enabled          bool                      `json:"enabled" boolean:"true"`
	DisabledString   int                       `json:"disabledString" string:"false"`
	Attempts         int                       `json:"attempts" required:"true" number:"true" min:"0"`
	DisabledBoolean  string                    `json:"disabledBoolean" boolean:"false"`
	Mode             string                    `json:"mode" required:"true" enum:"basic|strict"`
	DisabledNumber   string                    `json:"disabledNumber" number:"false"`
	Profile          ProbeValidationProfileDto  `json:"profile" required:"true" struct:"true"`
	DisabledStruct   string                    `json:"disabledStruct" struct:"false"`
	Labels           []ProbeValidationLabelDto  `json:"labels" required:"true" array:"true"`
	DisabledArray    string                    `json:"disabledArray" array:"false"`
}

type ProbeValidationProfileDto struct {
	Name string `json:"name" required:"true" string:"true"`
}

type ProbeValidationLabelDto struct {
	Code string `json:"code" required:"true" string:"true"`
}
