package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/application/types"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type GetCacheHashUseCase struct {
	redis ports.RedisPort
}

func NewGetCacheHashUseCase(redis ports.RedisPort) *GetCacheHashUseCase {
	return &GetCacheHashUseCase{redis: redis}
}

func (useCase *GetCacheHashUseCase) Execute(ctx context.Context, key string) (types.CacheOutput, error) {
	normalizedKey, value, err := useCase.redis.HGet(ctx, key, "value")
	return types.CacheOutput{Key: normalizedKey, Value: value}, err
}
