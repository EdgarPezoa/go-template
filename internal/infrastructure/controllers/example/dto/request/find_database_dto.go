package request

type FindDatabaseRequestDto struct {
	ID string `json:"-" required:"true" string:"true"`
}
