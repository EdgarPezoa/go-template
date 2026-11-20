//go:build integration

package mongodb

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

func TestMongoDBIntegration(t *testing.T) {
	host := os.Getenv("MONGODB_INTEGRATION_URL")
	if host == "" {
		t.Skip("MONGODB_INTEGRATION_URL is required")
	}
	port, err := strconv.Atoi(os.Getenv("MONGODB_INTEGRATION_PORT"))
	if err != nil || port == 0 {
		port = 27017
	}
	database := os.Getenv("MONGODB_INTEGRATION_DATABASE")
	if database == "" {
		database = "go-template-integration"
	}
	user, password, args := os.Getenv("MONGODB_INTEGRATION_USER"), os.Getenv("MONGODB_INTEGRATION_PASSWORD"), os.Getenv("MONGODB_INTEGRATION_ARGS")
	serverSelectionTimeoutSeconds, pingTimeoutSeconds := 3, 3
	if value := os.Getenv("MONGODB_INTEGRATION_SERVER_SELECTION_TIMEOUT_SECONDS"); value != "" {
		serverSelectionTimeoutSeconds, err = strconv.Atoi(value)
		if err != nil || serverSelectionTimeoutSeconds < 1 {
			t.Fatal("MONGODB_INTEGRATION_SERVER_SELECTION_TIMEOUT_SECONDS must be a positive integer")
		}
	}
	if value := os.Getenv("MONGODB_INTEGRATION_PING_TIMEOUT_SECONDS"); value != "" {
		pingTimeoutSeconds, err = strconv.Atoi(value)
		if err != nil || pingTimeoutSeconds < 1 {
			t.Fatal("MONGODB_INTEGRATION_PING_TIMEOUT_SECONDS must be a positive integer")
		}
	}
	adapter, err := NewMongoDBAdapter(host, user, password, args, database, port, serverSelectionTimeoutSeconds, pingTimeoutSeconds)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	value := "integration-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	collection := "example_values"
	t.Cleanup(func() {
		_, _ = adapter.database.Collection(collection).DeleteMany(context.Background(), bson.D{{Key: "value", Value: value}})
		_ = adapter.Disconnect(context.Background())
	})

	document := struct {
		Value string `bson:"value"`
	}{Value: value}
	inserted, err := adapter.Insert(ctx, collection, &document)
	if err != nil {
		t.Fatal(err)
	}
	var found struct {
		ID    string `bson:"_id,omitempty"`
		Value string `bson:"value"`
	}
	err = adapter.FindOne(ctx, collection, ports.IDFilter{ID: inserted}, &found)
	if err != nil || found.ID != inserted || found.Value != value {
		t.Fatalf("roundtrip = %#v, %v", found, err)
	}
	var values []struct {
		ID string `bson:"_id,omitempty"`
		Value string `bson:"value"`
	}
	if err := adapter.Find(ctx, collection, struct{ Value string `bson:"value"` }{Value: value}, &values); err != nil || len(values) != 1 {
		t.Fatalf("find = %#v, %v", values, err)
	}
	if err := adapter.Find(ctx, collection, struct{ Value string `bson:"value"` }{Value: "missing"}, &values); err != nil || len(values) != 0 {
		t.Fatalf("empty find = %#v, %v", values, err)
	}
	if err := adapter.FindOne(ctx, collection, ports.IDFilter{ID: bson.NewObjectID().Hex()}, &found); err != ports.ErrNotFound {
		t.Fatalf("missing error = %v, want not found", err)
	}
	if err := adapter.FindOne(ctx, collection, ports.IDFilter{ID: "invalid"}, &found); err != ports.ErrInvalidID {
		t.Fatalf("invalid error = %v, want invalid id", err)
	}
	upsertFilter := struct{ Value string `bson:"value"` }{Value: value + "-upsert"}
	upsertReplacement := struct {
		Value string `bson:"value"`
		Extra string `bson:"extra"`
	}{Value: upsertFilter.Value, Extra: "first"}
	upsert, err := adapter.Upsert(ctx, collection, upsertFilter, upsertReplacement)
	if err != nil || upsert.UpsertedID == "" || upsert.MatchedCount != 0 {
		t.Fatalf("upsert insert = %#v, %v", upsert, err)
	}
	replacement := struct {
		Value string `bson:"value"`
		Extra string `bson:"extra"`
	}{Value: upsertFilter.Value, Extra: "second"}
	upsert, err = adapter.Upsert(ctx, collection, upsertFilter, replacement)
	if err != nil || upsert.MatchedCount != 1 || upsert.UpsertedID != "" {
		t.Fatalf("upsert replacement = %#v, %v", upsert, err)
	}
	one, err := adapter.DeleteOne(ctx, collection, ports.IDFilter{ID: inserted})
	if err != nil || one.DeletedCount != 1 { t.Fatalf("delete one = %#v, %v", one, err) }
	many, err := adapter.DeleteMany(ctx, collection, struct{ Value string `bson:"value"` }{Value: upsertFilter.Value})
	if err != nil || many.DeletedCount != 1 { t.Fatalf("delete many = %#v, %v", many, err) }
}
