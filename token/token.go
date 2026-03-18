package token

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/silen/nuwa/cache/redis"
)

var ErrTokenNotFound = errors.New("token not found")

const defaultRedisDB = 3

func NewToken(ctx context.Context) *Token {
	return &Token{
		Ctx: ctx,
	}
}

type Token struct {
	Ctx context.Context
}

func (t *Token) Get(userInfo any, expiresIn time.Duration) (string, error) {
	rds, err := redis.NewRedis(t.Ctx, defaultRedisDB)
	if err != nil {
		return "", err
	}
	token := uuid.New().String()
	str, err := json.Marshal(userInfo)
	if err != nil {
		return "", err
	}
	err = rds.Set(t.Ctx, "token:"+token, string(str), expiresIn).Err()
	return token, err
}

func (t *Token) Check(token string) (res string, err error) {
	if token == "" {
		return "", ErrTokenNotFound
	}
	rds, err := redis.NewRedis(t.Ctx, defaultRedisDB)
	if err != nil {
		return
	}
	res, err = rds.Get(t.Ctx, "token:"+token).Result()
	if err == redis.Nil {
		return "", ErrTokenNotFound
	}
	return
}

func (t *Token) Del(token string) (res int64, err error) {
	rds, err := redis.NewRedis(t.Ctx, defaultRedisDB)
	if err != nil {
		return
	}
	res, err = rds.Del(t.Ctx, "token:"+token).Result()
	if err == redis.Nil {
		err = nil
	}
	return
}
