package request

type SendQueueRequestDto struct {
	Message string `json:"message" required:"true" string:"true"`
}
