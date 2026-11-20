package response

type SendTopicResponseDto struct {
	Accepted bool `json:"accepted" required:"true" boolean:"true"`
}
