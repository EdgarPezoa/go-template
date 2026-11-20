package request

type SendTopicRequestDto struct {
	Message string `json:"message" required:"true" string:"true"`
}
