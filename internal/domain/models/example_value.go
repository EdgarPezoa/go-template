package models

type ExampleValue struct {
	ID    string `bson:"_id,omitempty"`
	Value string
}
