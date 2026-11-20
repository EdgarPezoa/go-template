package request

type SetCacheHashRequestDto struct {
	Key   string `json:"-" required:"true" string:"true"`
	Value string `json:"value" required:"true" string:"true"`
}
