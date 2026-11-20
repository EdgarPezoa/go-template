package response

type HealthResponseDto struct {
	Status string `json:"status" required:"true" string:"true" enum:"ok"`
}
