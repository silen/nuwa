package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/logs"

	"github.com/redis/go-redis/v9"
)

// CacheInterface 定义缓存接口
type CacheInterface interface {
	Get(key string, dest interface{}) error
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(key string) error
	Exists(key string) (bool, error)
	Incr(key string) error
	Decr(key string) error
	Close() error
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
}

// GlobalCache 全局缓存实例
var GlobalCache CacheInterface

// InitCache 初始化缓存
func InitCache() {
	redisConf := conf.Redis
	if redisConf == nil || redisConf.Host == "" {
		logs.Warn("Redis configuration not found, cache functionality disabled")
		return
	}

	opts := &redis.Options{
		Addr:     redisConf.Host,
		Password: redisConf.Password,
		DB:       redisConf.DB,
		PoolSize: redisConf.PoolSize,
		MinIdleConns: redisConf.MinIdleConn,
		MaxIdleConns: redisConf.MaxIdleConn,
		ConnMaxLifetime: time.Duration(redisConf.MaxConnAge) * time.Second,
		DialTimeout: time.Duration(redisConf.DialTimeout) * time.Second,
		ReadTimeout: time.Duration(redisConf.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(redisConf.WriteTimeout) * time.Second,
		ConnMaxIdleTime: time.Duration(redisConf.IdleTimeout) * time.Second,
	}

	rdb := redis.NewClient(opts)

	cache := &RedisCache{
		client: rdb,
	}

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logs.Error("Failed to connect to Redis: ", err)
		return
	}

	GlobalCache = cache
	logs.Info("Redis cache initialized successfully")
}

// Get 从缓存获取数据
func (r *RedisCache) Get(key string, dest interface{}) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s does not exist", key)
		}
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set 设置缓存数据
func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Delete 删除缓存数据
func (r *RedisCache) Delete(key string) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

// Exists 检查缓存键是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	if r.client == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Incr 增加计数器
func (r *RedisCache) Incr(key string) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	return r.client.Incr(ctx, key).Err()
}

// Decr 减少计数器
func (r *RedisCache) Decr(key string) error {
	if r.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	ctx := context.Background()
	return r.client.Decr(ctx, key).Err()
}

// Close 关闭缓存连接
func (r *RedisCache) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}

// GetCache 获取全局缓存实例
func GetCache() CacheInterface {
	return GlobalCache
}

// 缓存辅助函数

// GetOrSet 尝试从缓存获取数据，如果不存在则执行fn并缓存结果
func GetOrSet(key string, dest interface{}, expiration time.Duration, fn func() (interface{}, error)) error {
	cache := GetCache()
	if cache == nil {
		// 如果没有缓存，直接执行fn
		result, err := fn()
		if err != nil {
			return err
		}
		// 尝试转换result到dest类型
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, dest)
	}

	// 尝试从缓存获取
	if err := cache.Get(key, dest); err != nil {
		// 缓存未命中，执行fn并设置缓存
		result, err := fn()
		if err != nil {
			return err
		}

		if err := cache.Set(key, result, expiration); err != nil {
			logs.Error("Failed to set cache: ", err)
		}

		// 将结果复制到dest
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, dest)
	}

	return nil
}