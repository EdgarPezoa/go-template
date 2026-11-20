package request

type GetCacheHashRequestDto struct {
	Key string `json:"-" required:"true" string:"true"`
}
