package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s Storage) SetCache(key string, value interface{}, exp time.Duration) {
	s.rds.Set(context.Background(), key, value, exp)
}

func (s Storage) GetCache(key string) *redis.StringCmd {
	return s.rds.Get(context.Background(), key)
}
