package request

type GetCacheRequestDto struct {
	Key string `json:"-" required:"true" string:"true"`
}
