package redis

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type RedisAdapter struct {
	address  string
	password string
	db       int
	timeout  time.Duration
}

func NewRedisAdapter(address, password string, database int) (*RedisAdapter, error) {
	if address == "" {
		return nil, errors.New("redis address is required")
	}
	redis := &RedisAdapter{
		address:  address,
		password: password,
		db:       database,
		timeout:  3 * time.Second,
	}
	connection, err := redis.conn(context.Background())
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	if _, err := redis.command(connection, "PING"); err != nil {
		return nil, err
	}
	return redis, nil
}

func (redis *RedisAdapter) conn(ctx context.Context) (net.Conn, error) {
	dialer := net.Dialer{Timeout: redis.timeout}
	connection, err := dialer.DialContext(ctx, "tcp", redis.address)
	if err != nil {
		return nil, err
	}
	if redis.password != "" {
		if _, err := redis.command(connection, "AUTH", redis.password); err != nil {
			connection.Close()
			return nil, err
		}
	}
	if redis.db != 0 {
		if _, err := redis.command(connection, "SELECT", strconv.Itoa(redis.db)); err != nil {
			connection.Close()
			return nil, err
		}
	}
	return connection, nil
}

func (redis *RedisAdapter) exec(ctx context.Context, arguments ...string) (any, error) {
	connection, err := redis.conn(ctx)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetDeadline(deadline)
	}
	return redis.command(connection, arguments...)
}

func (redis *RedisAdapter) command(connection net.Conn, arguments ...string) (any, error) {
	var request strings.Builder
	fmt.Fprintf(&request, "*%d\r\n", len(arguments))
	for _, arg := range arguments {
		fmt.Fprintf(&request, "$%d\r\n%s\r\n", len(arg), arg)
	}
	if _, err := connection.Write([]byte(request.String())); err != nil {
		return nil, err
	}
	return readRESP(bufio.NewReader(connection))
}

func (redis *RedisAdapter) Get(ctx context.Context, key string) (string, string, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return "", "", err
	}
	value, err := redis.exec(ctx, "GET", normalizedKey)
	stringValue, err := valueString(value, err)
	return normalizedKey, stringValue, err
}

func (redis *RedisAdapter) Set(ctx context.Context, key, value string) (string, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	_, err = redis.exec(ctx, "SET", normalizedKey, value)
	return normalizedKey, err
}

func (redis *RedisAdapter) HGet(ctx context.Context, key, field string) (string, string, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return "", "", err
	}
	value, err := redis.exec(ctx, "HGET", normalizedKey, field)
	stringValue, err := valueString(value, err)
	return normalizedKey, stringValue, err
}

func (redis *RedisAdapter) HSet(ctx context.Context, key, field, value string) (string, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return "", err
	}
	_, err = redis.exec(ctx, "HSET", normalizedKey, field, value)
	return normalizedKey, err
}

func (redis *RedisAdapter) Close() error {
	return nil
}

func normalizeKey(key string) (string, error) {
	normalizedKey := strings.TrimSpace(key)
	if normalizedKey == "" {
		return "", ports.ErrInvalidKey
	}
	return normalizedKey, nil
}

func valueString(value any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", ports.ErrNotFound
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("unexpected redis response")
	}
	return stringValue, nil
}

func readRESP(reader *bufio.Reader) (any, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")

	switch prefix {
	case '+':
		return line, nil
	case '-':
		return nil, errors.New(line)
	case ':':
		return line, nil
	case '$':
		length, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid RESP bulk length %q: %w", line, err)
		}
		if length < 0 {
			return nil, nil
		}
		payload := make([]byte, length+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return nil, err
		}
		if payload[length] != '\r' || payload[length+1] != '\n' {
			return nil, errors.New("invalid RESP bulk payload terminator")
		}
		return string(payload[:length]), nil
	default:
		return nil, fmt.Errorf("unsupported redis response %q", prefix)
	}
}
