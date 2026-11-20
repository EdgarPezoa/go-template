package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/application/types"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type GetCacheUseCase struct {
	redis ports.RedisPort
}

func NewGetCacheUseCase(redis ports.RedisPort) *GetCacheUseCase {
	return &GetCacheUseCase{redis: redis}
}

func (useCase *GetCacheUseCase) Execute(ctx context.Context, key string) (types.CacheOutput, error) {
	normalizedKey, value, err := useCase.redis.Get(ctx, key)
	return types.CacheOutput{Key: normalizedKey, Value: value}, err
}
