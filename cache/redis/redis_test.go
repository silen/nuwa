package redis

import (
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestRedisClientKeyIncludesDB(t *testing.T) {
	t.Parallel()

	key1 := redisClientKey(&goredis.Options{Addr: "127.0.0.1:6379", Password: "pw", DB: 1})
	key2 := redisClientKey(&goredis.Options{Addr: "127.0.0.1:6379", Password: "pw", DB: 2})

	if key1 == key2 {
		t.Fatalf("expected different redis client keys for different DBs")
	}
}
