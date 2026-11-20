package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/domain/models"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type InsertDatabaseUseCase struct {
	database ports.DatabasePort
}

func NewInsertDatabaseUseCase(database ports.DatabasePort) *InsertDatabaseUseCase {
	return &InsertDatabaseUseCase{database: database}
}

func (useCase *InsertDatabaseUseCase) Execute(ctx context.Context, input models.ExampleValue) (models.ExampleValue, error) {
	const collection = "example_values"
	id, err := useCase.database.Insert(ctx, collection, &input)
	if err != nil {
		return models.ExampleValue{}, err
	}
	input.ID = id
	return input, nil
}
