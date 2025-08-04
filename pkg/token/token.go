package token

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/silen/nuwa/pkg/conf"
	"github.com/silen/nuwa/pkg/redis"
)

func NewToken(ctx context.Context) *Token {
	return &Token{
		Ctx: ctx,
	}
}

type Token struct {
	Ctx context.Context
}

func (t *Token) Get(userInfo any, expiresIn time.Duration) (string, error) {
	rds, err := redis.NewRedis(t.Ctx, conf.TOKEN_BY_RDS_BD)
	if err != nil {
		return "", err
	}
	token := uuid.New().String()
	str, _ := json.Marshal(userInfo)
	err = rds.Set(t.Ctx, "token:"+token, string(str), expiresIn).Err()
	return token, err
}

func (t *Token) Check(token string) (res string, err error) {
	rds, err := redis.NewRedis(t.Ctx, conf.TOKEN_BY_RDS_BD)
	if err != nil {
		return
	}
	res, err = rds.Get(t.Ctx, "token:"+token).Result()
	if err == redis.Nil {
		err = nil
	}
	return
}

func (t *Token) Del(token string) (res int64, err error) {
	rds, err := redis.NewRedis(t.Ctx, conf.TOKEN_BY_RDS_BD)
	if err != nil {
		return
	}
	res, err = rds.Del(t.Ctx, "token:"+token).Result()
	if err == redis.Nil {
		err = nil
	}
	return
}
