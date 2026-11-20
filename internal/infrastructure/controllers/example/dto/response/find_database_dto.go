package response

type FindDatabaseResponseDto struct {
	ID    string `json:"id" required:"true" string:"true"`
	Value string `json:"value" required:"true" string:"true"`
}
