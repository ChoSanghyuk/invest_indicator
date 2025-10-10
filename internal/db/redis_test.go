package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {

	rc := NewRedisConfig("", "127.0.0.1", "6379", 0)
	rds := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", rc.ip, rc.port),
		Password: rc.password, //
		DB:       rc.db,       //
	})

	t.Run("setAndGetBoolValue", func(t *testing.T) {
		key := "test1"
		value := true
		exp := time.Hour

		rds.Set(context.Background(), key, value, exp)
		rtnCmd := rds.Get(context.Background(), key)

		rtnValue, err := rtnCmd.Bool()

		assert.NoError(t, err)
		assert.Equal(t, value, rtnValue)
	})

	t.Run("GetBoolValue", func(t *testing.T) {
		key := "test1"
		value := true

		rtnCmd := rds.Get(context.Background(), key)

		rtnValue, err := rtnCmd.Bool()

		assert.NoError(t, err)
		assert.Equal(t, value, rtnValue)
	})

	t.Run("GetBoolValue", func(t *testing.T) {
		key := "test2"
		value := true

		rtnCmd := rds.Get(context.Background(), key)

		rtnValue, err := rtnCmd.Bool()

		assert.NoError(t, err)
		assert.Equal(t, !value, rtnValue)
	})
}
