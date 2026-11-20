package response

type ReadyResponseDto struct {
	Status string `json:"status" required:"true" string:"true" enum:"ready|not_ready"`
}
