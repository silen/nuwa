package redis

import (
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"
)

var (
	client *redis.Client
)

const Nil = redis.Nil

func NewRedis(ctx context.Context, db ...int) (*redis.Client, error) {

	op := &redis.Options{
		Addr:     conf.Redis.Host,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	}

	if len(db) > 0 {
		op.DB = db[0]
	}

	if client != nil {
		if _, err := client.Ping(ctx).Result(); err == nil {
			client.Do(ctx, "select", op.DB)
			return client, nil
		}
	}

	client = redis.NewClient(op)

	if _, err := client.Ping(ctx).Result(); err != nil {
		logs.WithContext(ctx).Error("redis conn error:", err.Error())
		return nil, err
	}
	return client, nil
}
