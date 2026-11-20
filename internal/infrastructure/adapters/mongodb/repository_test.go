package mongodb

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/edgarpezoa/go-template/internal/domain/models"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

func TestExampleValueDocumentMapping(t *testing.T) {
	id, err := bson.ObjectIDFromHex("507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := bson.Marshal(bson.D{{Key: "_id", Value: id}, {Key: "value", Value: "value"}})
	if err != nil {
		t.Fatal(err)
	}
	var result models.ExampleValue
	decoder := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(raw)))
	decoder.ObjectIDAsHexString()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result != (models.ExampleValue{ID: id.Hex(), Value: "value"}) {
		t.Fatalf("result = %#v", result)
	}
}

func TestIDFilterRejectsInvalidIDs(t *testing.T) {
	for _, filter := range []any{ports.IDFilter{ID: "invalid"}, (*ports.IDFilter)(nil)} {
		if _, err := normalizeFilter(filter); !errors.Is(err, ports.ErrInvalidID) {
			t.Fatalf("filter %#v error = %v, want invalid id", filter, err)
		}
	}
}

func TestIDFilterMapsToObjectID(t *testing.T) {
	id, err := bson.ObjectIDFromHex("507f1f77bcf86cd799439011")
	if err != nil {
		t.Fatal(err)
	}
	filter, err := normalizeFilter(ports.IDFilter{ID: id.Hex()})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(filter, bson.D{{Key: "_id", Value: id}}) {
		t.Fatalf("filter = %#v", filter)
	}
}

func TestRepositoryValidatesGenericArguments(t *testing.T) {
	adapter := &MongoDBAdapter{}
	ctx := context.Background()
	var document models.ExampleValue
	var results []models.ExampleValue
	if _, err := adapter.Insert(ctx, "", document); err == nil { t.Fatal("expected empty collection error") }
	if _, err := adapter.Insert(ctx, "example_values", nil); err == nil { t.Fatal("expected nil document error") }
	if err := adapter.FindOne(ctx, "", bson.D{}, &document); err == nil { t.Fatal("expected empty collection error") }
	if err := adapter.FindOne(ctx, "example_values", bson.D{}, nil); err == nil { t.Fatal("expected nil result error") }
	if err := adapter.FindOne(ctx, "example_values", bson.D{}, document); err == nil { t.Fatal("expected non-pointer result error") }
	if err := adapter.Find(ctx, "example_values", bson.D{}, results); err == nil { t.Fatal("expected non-pointer slice result error") }
	if err := adapter.Find(ctx, "example_values", bson.D{}, &document); err == nil { t.Fatal("expected non-slice result error") }
	if _, err := adapter.Upsert(ctx, "example_values", bson.D{}, document); err == nil { t.Fatal("expected empty filter error") }
	if _, err := adapter.Upsert(ctx, "example_values", bson.D{{Key: "value", Value: "v"}}, nil); err == nil { t.Fatal("expected empty replacement error") }
}

func TestExampleValueRepositoryErrorsRemainGeneric(t *testing.T) {
	if !errors.Is(ports.ErrInvalidID, ports.ErrInvalidID) || !errors.Is(ports.ErrNotFound, ports.ErrNotFound) {
		t.Fatal("expected generic database errors")
	}
}
