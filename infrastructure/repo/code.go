package repo

import (
	"CloudDrive/domain/account/repository"
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

type codeRepo struct {
	rdb *redis.Client
	ctx context.Context
}

func (cr *codeRepo) SetCode(codeKey, codeValue string, expiration time.Duration) error {
	if err := cr.rdb.Set(cr.ctx, codeKey, codeValue, expiration).Err(); err != nil {
		return err
	}
	return nil
}

func (cr *codeRepo) GetCode(codeKey string) (string, error) {
	code, err := cr.rdb.Get(cr.ctx, codeKey).Result()
	if err != nil {
		return "", err
	}
	return code, nil
}

func NewCodeRepo(rdb *redis.Client, ctx context.Context) (repository.CodeRepository, error) {
	if rdb == nil {
		return nil, fmt.Errorf("missing redis client")
	}
	return &codeRepo{
		rdb: rdb,
		ctx: ctx,
	}, nil
}
