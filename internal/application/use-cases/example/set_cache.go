package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/application/types"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type SetCacheUseCase struct {
	redis ports.RedisPort
}

func NewSetCacheUseCase(redis ports.RedisPort) *SetCacheUseCase {
	return &SetCacheUseCase{redis: redis}
}

func (useCase *SetCacheUseCase) Execute(ctx context.Context, input types.CacheInput) (types.CacheOutput, error) {
	normalizedKey, err := useCase.redis.Set(ctx, input.Key, input.Value)
	return types.CacheOutput{Key: normalizedKey, Value: input.Value}, err
}
