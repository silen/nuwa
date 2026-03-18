package redis

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/silen/nuwa/config"
	"github.com/silen/nuwa/logs"
)

var (
	clientsMu sync.Mutex
	clients   = make(map[string]*redis.Client)
)

const Nil = redis.Nil

func NewRedis(ctx context.Context, db ...int) (*redis.Client, error) {
	if config.Redis == nil {
		return nil, errors.New("redis config not initialized")
	}

	op := &redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	}

	if len(db) > 0 {
		op.DB = db[0]
	}

	clientKey := redisClientKey(op)

	clientsMu.Lock()
	defer clientsMu.Unlock()

	if client, ok := clients[clientKey]; ok {
		if _, err := client.Ping(ctx).Result(); err == nil {
			return client, nil
		}
		_ = client.Close()
		delete(clients, clientKey)
	}

	client := redis.NewClient(op)

	if _, err := client.Ping(ctx).Result(); err != nil {
		logs.WithContext(ctx).Error("redis conn error:", err.Error())
		return nil, err
	}
	clients[clientKey] = client
	return client, nil
}

func redisClientKey(op *redis.Options) string {
	return op.Addr + "|" + op.Password + "|" + redisDBString(op.DB)
}

func redisDBString(db int) string {
	return strconv.Itoa(db)
}
