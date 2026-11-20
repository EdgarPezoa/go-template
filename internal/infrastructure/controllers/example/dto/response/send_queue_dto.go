package response

type SendQueueResponseDto struct {
	Accepted bool `json:"accepted" required:"true" boolean:"true"`
}
