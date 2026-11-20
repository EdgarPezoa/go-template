package example

import (
	"context"
	"errors"
	"testing"

	"github.com/edgarpezoa/go-template/internal/domain/models"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type databasePortFake struct {
	document any
	collection string
	findFilter any
	result any
	findErr error
	insertErr error
}

func (fake *databasePortFake) DeleteOne(context.Context, string, any) (ports.DeleteResult, error) { return ports.DeleteResult{}, nil }
func (fake *databasePortFake) DeleteMany(context.Context, string, any) (ports.DeleteResult, error) { return ports.DeleteResult{}, nil }
func (fake *databasePortFake) FindOne(_ context.Context, collection string, filter, result any) error {
	fake.collection = collection
	fake.findFilter = filter
	fake.result = result
	if fake.findErr != nil { return fake.findErr }
	*result.(*models.ExampleValue) = models.ExampleValue{ID: "id", Value: "value"}
	return nil
}
func (fake *databasePortFake) Find(context.Context, string, any, any) error { return nil }
func (fake *databasePortFake) Insert(_ context.Context, collection string, document any) (string, error) {
	fake.collection = collection
	fake.document = document
	if fake.insertErr != nil { return "", fake.insertErr }
	return "id", nil
}
func (fake *databasePortFake) Upsert(context.Context, string, any, any) (ports.UpsertResult, error) { return ports.UpsertResult{}, nil }

func TestDatabaseUseCases(t *testing.T) {
	port := &databasePortFake{}
	inserted, err := NewInsertDatabaseUseCase(port).Execute(context.Background(), models.ExampleValue{Value: "value"})
	if err != nil { t.Fatalf("insert error = %v", err) }
	if inserted.ID != "id" || inserted.Value != "value" { t.Fatalf("insert = %+v, %v", inserted, err) }
	if port.collection != "example_values" || port.document.(*models.ExampleValue).Value != "value" { t.Fatalf("insert forwarding = %#v", port.document) }
	output, err := NewFindDatabaseUseCase(port).Execute(context.Background(), "id")
	if err != nil { t.Fatalf("find error = %v", err) }
	if output.ID != "id" || output.Value != "value" { t.Fatalf("find = %+v, %v", output, err) }
	if port.collection != "example_values" || port.findFilter != (ports.IDFilter{ID: "id"}) || port.result == nil { t.Fatalf("find forwarding = %#v, %#v", port.findFilter, port.result) }
	port.insertErr = errors.New("insert failed")
	if _, err := NewInsertDatabaseUseCase(port).Execute(context.Background(), models.ExampleValue{Value: "value"}); !errors.Is(err, port.insertErr) { t.Fatalf("insert error = %v, want propagated error", err) }
	port.findErr = ports.ErrNotFound
	if _, err := NewFindDatabaseUseCase(port).Execute(context.Background(), "missing"); !errors.Is(err, ports.ErrNotFound) { t.Fatalf("find error = %v, want not found", err) }
}
