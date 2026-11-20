package response

type InsertDatabaseResponseDto struct {
	ID    string `json:"id" required:"true" string:"true"`
	Value string `json:"value" required:"true" string:"true"`
}
