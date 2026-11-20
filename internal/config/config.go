package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/edgarpezoa/go-template/internal/shared/validation"
)

type Config struct {
	API        APIConfig
	Redis      RedisConfig
	ServiceBus ServiceBusConfig
	Mongo      MongoConfig
}

type APIConfig struct {
	Port int `env:"API_PORT" min:"1" max:"65535"`
}

type RedisConfig struct {
	Address  string `env:"REDIS_ADDRESS"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" default:"0" min:"0"`
}

type ServiceBusConfig struct {
	ConnectionString string `env:"AZURE_SERVICEBUS_CONNECTION_STRING"`
	Queue            string `env:"AZURE_SERVICEBUS_QUEUE" default:"go-template-messages"`
	Topic            string `env:"AZURE_SERVICEBUS_TOPIC" default:"go-template-events"`
	Subscription     string `env:"AZURE_SERVICEBUS_SUBSCRIPTION" default:"go-template-worker"`
}

type MongoConfig struct {
	URL                          string `env:"MONGODB_URL"`
	User                         string `env:"MONGODB_USER"`
	Password                     string `env:"MONGODB_PASSWORD"`
	Args                         string `env:"MONGODB_ARGS"`
	Database                     string `env:"MONGODB_DATABASE"`
	Port                         int    `env:"MONGODB_PORT" min:"1" max:"65535"`
	ServerSelectionTimeoutSeconds int    `env:"MONGODB_SERVER_SELECTION_TIMEOUT_SECONDS" min:"1"`
	PingTimeoutSeconds             int    `env:"MONGODB_PING_TIMEOUT_SECONDS" min:"1"`
}

func Load() (Config, error) {
	var cfg Config
	dotEnvContents, err := readDotEnv()
	if err != nil {
		return Config{}, err
	}
	for _, line := range strings.Split(dotEnvContents, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		if err := os.Setenv(key, strings.TrimSpace(value)); err != nil {
			return Config{}, fmt.Errorf("set environment variable %q: %w", key, err)
		}
	}

	cfg.API.Port, err = envInt("API_PORT")
	if err != nil {
		return Config{}, err
	}
	if os.Getenv("API_PORT") == "" {
		cfg.API.Port = 3000
	}
	cfg.Redis.Address = os.Getenv("REDIS_ADDRESS")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.DB, err = envInt("REDIS_DB")
	if err != nil {
		return Config{}, err
	}
	validator := validation.NewValidator()
	cfg.ServiceBus.ConnectionString = os.Getenv("AZURE_SERVICEBUS_CONNECTION_STRING")
	cfg.ServiceBus.Queue = os.Getenv("AZURE_SERVICEBUS_QUEUE")
	cfg.ServiceBus.Topic = os.Getenv("AZURE_SERVICEBUS_TOPIC")
	cfg.ServiceBus.Subscription = os.Getenv("AZURE_SERVICEBUS_SUBSCRIPTION")
	cfg.Mongo.URL = os.Getenv("MONGODB_URL")
	cfg.Mongo.User = os.Getenv("MONGODB_USER")
	cfg.Mongo.Password = os.Getenv("MONGODB_PASSWORD")
	cfg.Mongo.Args, err = canonicalQuery("MONGODB_ARGS")
	if err != nil {
		return Config{}, err
	}
	cfg.Mongo.Database = os.Getenv("MONGODB_DATABASE")
	cfg.Mongo.Port, err = envInt("MONGODB_PORT")
	if err != nil {
		return Config{}, err
	}
	if cfg.Mongo.URL == "" {
		cfg.Mongo.URL = "localhost"
	}
	if cfg.Mongo.Port == 0 {
		cfg.Mongo.Port = 27017
	}
	cfg.Mongo.ServerSelectionTimeoutSeconds, err = envInt("MONGODB_SERVER_SELECTION_TIMEOUT_SECONDS")
	if err != nil {
		return Config{}, err
	}
	if os.Getenv("MONGODB_SERVER_SELECTION_TIMEOUT_SECONDS") == "" {
		cfg.Mongo.ServerSelectionTimeoutSeconds = 3
	} else if cfg.Mongo.ServerSelectionTimeoutSeconds <= 0 {
		return Config{}, fmt.Errorf("config validation: MONGODB_SERVER_SELECTION_TIMEOUT_SECONDS must be greater than zero")
	}
	cfg.Mongo.PingTimeoutSeconds, err = envInt("MONGODB_PING_TIMEOUT_SECONDS")
	if err != nil {
		return Config{}, err
	}
	if os.Getenv("MONGODB_PING_TIMEOUT_SECONDS") == "" {
		cfg.Mongo.PingTimeoutSeconds = 3
	} else if cfg.Mongo.PingTimeoutSeconds <= 0 {
		return Config{}, fmt.Errorf("config validation: MONGODB_PING_TIMEOUT_SECONDS must be greater than zero")
	}
	cfg.Mongo.Database = strings.TrimSpace(cfg.Mongo.Database)
	if cfg.Mongo.Database == "" {
		cfg.Mongo.Database = "go-template"
	}
	if strings.Contains(cfg.Mongo.URL, "://") || strings.ContainsAny(cfg.Mongo.URL, ":/?#@") {
		return Config{}, fmt.Errorf("config validation: MONGODB_URL must be a host without scheme, credentials, port, path, or query")
	}
	if (cfg.Mongo.User == "") != (cfg.Mongo.Password == "") {
		return Config{}, fmt.Errorf("config validation: MONGODB_USER and MONGODB_PASSWORD must be provided together")
	}
	validator.ApplyDefaults(&cfg)
	if err := validator.Validate(cfg); err != nil {
		return Config{}, fmt.Errorf("config validation: %w", err)
	}
	return cfg, nil
}

func readDotEnv() (string, error) {
	bytes, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read .env: %w", err)
	}
	return string(bytes), nil
}

func envInt(name string) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return 0, nil
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as integer: %w", name, err)
	}
	return number, nil
}

func canonicalQuery(name string) (string, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return "", nil
	}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return "", fmt.Errorf("config validation: %s must be valid URL query arguments: %w", name, err)
	}
	return values.Encode(), nil
}
