package request

type InsertDatabaseRequestDto struct {
	Value string `json:"value" required:"true" string:"true"`
}
