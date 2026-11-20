package response

type GetCacheResponseDto struct {
	Key   string `json:"key" required:"true" string:"true"`
	Value string `json:"value" required:"true" string:"true"`
}
