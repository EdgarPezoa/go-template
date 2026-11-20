package response

type SetCacheHashResponseDto struct {
	Key   string `json:"key" required:"true" string:"true"`
	Value string `json:"value" required:"true" string:"true"`
}
