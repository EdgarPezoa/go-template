package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/domain/models"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type FindDatabaseUseCase struct {
	database ports.DatabasePort
}

func NewFindDatabaseUseCase(database ports.DatabasePort) *FindDatabaseUseCase {
	return &FindDatabaseUseCase{database: database}
}

func (useCase *FindDatabaseUseCase) Execute(ctx context.Context, id string) (models.ExampleValue, error) {
	const collection = "example_values"
	var result models.ExampleValue
	if err := useCase.database.FindOne(ctx, collection, ports.IDFilter{ID: id}, &result); err != nil {
		return models.ExampleValue{}, err
	}
	return result, nil
}
