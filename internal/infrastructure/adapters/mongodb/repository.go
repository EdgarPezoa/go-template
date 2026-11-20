package mongodb

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

func (adapter *MongoDBAdapter) DeleteOne(ctx context.Context, collection string, filter any) (ports.DeleteResult, error) {
	if err := validateCollection(collection); err != nil {
		return ports.DeleteResult{}, err
	}
	filter, err := normalizeFilter(filter)
	if err != nil {
		return ports.DeleteResult{}, err
	}
	result, err := adapter.database.Collection(collection).DeleteOne(ctx, filter)
	if err != nil {
		return ports.DeleteResult{}, err
	}
	return ports.DeleteResult{DeletedCount: result.DeletedCount}, nil
}

func (adapter *MongoDBAdapter) DeleteMany(ctx context.Context, collection string, filter any) (ports.DeleteResult, error) {
	if err := validateCollection(collection); err != nil {
		return ports.DeleteResult{}, err
	}
	filter, err := normalizeFilter(filter)
	if err != nil {
		return ports.DeleteResult{}, err
	}
	result, err := adapter.database.Collection(collection).DeleteMany(ctx, filter)
	if err != nil {
		return ports.DeleteResult{}, err
	}
	return ports.DeleteResult{DeletedCount: result.DeletedCount}, nil
}

func (adapter *MongoDBAdapter) FindOne(ctx context.Context, collection string, filter, result any) error {
	if err := validateCollection(collection); err != nil {
		return err
	}
	if isNil(result) || reflect.ValueOf(result).Kind() != reflect.Pointer {
		return errors.New("mongodb result must be a non-nil pointer")
	}
	filter, err := normalizeFilter(filter)
	if err != nil {
		return err
	}
	err = adapter.database.Collection(collection).FindOne(ctx, filter).Decode(result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ports.ErrNotFound
	}
	return err
}

func (adapter *MongoDBAdapter) Find(ctx context.Context, collection string, filter, results any) error {
	if err := validateCollection(collection); err != nil {
		return err
	}
	if isNil(results) || reflect.ValueOf(results).Kind() != reflect.Pointer || reflect.ValueOf(results).Elem().Kind() != reflect.Slice {
		return errors.New("mongodb results must be a non-nil pointer to a slice")
	}
	filter, err := normalizeFilter(filter)
	if err != nil {
		return err
	}
	cursor, err := adapter.database.Collection(collection).Find(ctx, filter)
	if err != nil {
		return err
	}
	if err := cursor.All(ctx, results); err != nil {
		_ = cursor.Close(ctx)
		return err
	}
	if err := cursor.Close(ctx); err != nil {
		return err
	}
	return cursor.Err()
}

func (adapter *MongoDBAdapter) Insert(ctx context.Context, collection string, document any) (string, error) {
	if err := validateCollection(collection); err != nil {
		return "", err
	}
	if isNil(document) {
		return "", errors.New("mongodb document is required")
	}
	result, err := adapter.database.Collection(collection).InsertOne(ctx, document)
	if err != nil {
		return "", err
	}
	id, ok := result.InsertedID.(bson.ObjectID)
	if !ok {
		return "", errors.New("mongodb returned a non-object id")
	}
	return id.Hex(), nil
}

func (adapter *MongoDBAdapter) Upsert(ctx context.Context, collection string, filter, replacement any) (ports.UpsertResult, error) {
	if err := validateCollection(collection); err != nil {
		return ports.UpsertResult{}, err
	}
	if isEmpty(filter) {
		return ports.UpsertResult{}, errors.New("mongodb upsert filter is required")
	}
	if isEmpty(replacement) {
		return ports.UpsertResult{}, errors.New("mongodb upsert replacement is required")
	}
	filter, err := normalizeFilter(filter)
	if err != nil {
		return ports.UpsertResult{}, err
	}
	result, err := adapter.database.Collection(collection).ReplaceOne(ctx, filter, replacement, options.Replace().SetUpsert(true))
	if err != nil {
		return ports.UpsertResult{}, err
	}
	output := ports.UpsertResult{MatchedCount: result.MatchedCount, ModifiedCount: result.ModifiedCount}
	if result.UpsertedID != nil {
		id, ok := result.UpsertedID.(bson.ObjectID)
		if !ok {
			return ports.UpsertResult{}, errors.New("mongodb returned a non-object upserted id")
		}
		output.UpsertedID = id.Hex()
	}
	return output, nil
}

func validateCollection(collection string) error {
	if strings.TrimSpace(collection) == "" {
		return errors.New("mongodb collection is required")
	}
	return nil
}

func normalizeFilter(filter any) (any, error) {
	switch value := filter.(type) {
	case ports.IDFilter:
		id, err := bson.ObjectIDFromHex(value.ID)
		if err != nil {
			return nil, ports.ErrInvalidID
		}
		return bson.D{{Key: "_id", Value: id}}, nil
	case *ports.IDFilter:
		if value == nil {
			return nil, ports.ErrInvalidID
		}
		id, err := bson.ObjectIDFromHex(value.ID)
		if err != nil {
			return nil, ports.ErrInvalidID
		}
		return bson.D{{Key: "_id", Value: id}}, nil
	default:
		return filter, nil
	}
}

func isEmpty(value any) bool {
	if isNil(value) {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		return v.IsZero()
	default:
		return false
	}
}

func isNil(value any) bool {
	if value == nil {
		return true
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflect.ValueOf(value).IsNil()
	default:
		return false
	}
}
