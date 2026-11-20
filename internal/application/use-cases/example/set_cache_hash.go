package example

import (
	"context"

	"github.com/edgarpezoa/go-template/internal/application/types"
	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type SetCacheHashUseCase struct {
	redis ports.RedisPort
}

func NewSetCacheHashUseCase(redis ports.RedisPort) *SetCacheHashUseCase {
	return &SetCacheHashUseCase{redis: redis}
}

func (useCase *SetCacheHashUseCase) Execute(ctx context.Context, input types.CacheInput) (types.CacheOutput, error) {
	normalizedKey, err := useCase.redis.HSet(ctx, input.Key, "value", input.Value)
	return types.CacheOutput{Key: normalizedKey, Value: input.Value}, err
}
