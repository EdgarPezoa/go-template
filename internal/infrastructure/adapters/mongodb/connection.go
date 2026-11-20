package mongodb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDBAdapter struct {
	client   *mongo.Client
	database *mongo.Database
}

func NewMongoDBAdapter(host, user, password, args, database string, port, serverSelectionTimeoutSeconds, pingTimeoutSeconds int) (*MongoDBAdapter, error) {
	uri, err := buildURI(host, user, password, args, database, port)
	if err != nil {
		return nil, err
	}
	if serverSelectionTimeoutSeconds < 1 || pingTimeoutSeconds < 1 {
		return nil, errors.New("mongodb timeouts must be positive")
	}
	clientOptions := options.Client().ApplyURI(uri).
		SetBSONOptions(&options.BSONOptions{ObjectIDAsHexString: true}).
		SetServerSelectionTimeout(time.Duration(serverSelectionTimeoutSeconds) * time.Second)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pingTimeoutSeconds)*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongodb: %w", err)
	}
	return &MongoDBAdapter{
		client:   client,
		database: client.Database(database),
	}, nil
}

func buildURI(host, user, password, args, database string, port int) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "localhost"
	}
	if strings.Contains(host, "://") || strings.ContainsAny(host, ":/?#@") {
		return "", errors.New("mongodb host must not contain scheme, path, query, fragment, or credentials")
	}
	if database == "" {
		return "", errors.New("mongodb database is required")
	}
	if port < 1 || port > 65535 {
		return "", errors.New("mongodb port must be between 1 and 65535")
	}
	credentials := ""
	if user != "" || password != "" {
		if user == "" || password == "" {
			return "", errors.New("mongodb user and password must be provided together")
		}
		credentials = url.UserPassword(user, password).String() + "@"
	}
	uri := "mongodb://" + credentials + net.JoinHostPort(host, strconv.Itoa(port)) + "/" + url.PathEscape(database)
	if args != "" {
		query, err := url.ParseQuery(args)
		if err != nil {
			return "", fmt.Errorf("mongodb args must be valid URL query arguments: %w", err)
		}
		uri += "?" + query.Encode()
	}
	return uri, nil
}

func (adapter *MongoDBAdapter) Disconnect(ctx context.Context) error {
	if adapter == nil || adapter.client == nil {
		return nil
	}
	return adapter.client.Disconnect(ctx)
}
