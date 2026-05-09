package limiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/fixed_window.lua
var fixedWindowScript string

type RedisLimiter struct {
	client *redis.Client
	script *redis.Script
}

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		script: redis.NewScript(fixedWindowScript),
	}
}

func (r *RedisLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (Result, error) {
	res, err := r.script.Run(ctx, r.client, []string{key}, limit, window.Milliseconds()).Int64Slice()
	if err != nil {
		return Result{}, err
	}

	allowed := res[0] == 1
	remaining := int(res[1])
	return Result{
		Allowed:   allowed,
		Remaining: remaining,
		ResetTime: time.Now().Add(window)}, nil
}
