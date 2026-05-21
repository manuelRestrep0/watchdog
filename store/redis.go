package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manuelRestrep0/watchdog/model"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", &err)
	}

	return &RedisStore{client: client}, nil
}

func (r *RedisStore) SetLastCheck(ctx context.Context, check *model.Check) error {
	data, err := json.Marshal(check)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("target:%d:last_check", check.TargetID)
	return r.client.Set(ctx, key, data, 2*time.Hour).Err()
}

func (r *RedisStore) GetLastCheck(ctx context.Context, targetID int64) (*model.Check, error) {
	key := fmt.Sprintf("target:%d:last_ccheck", targetID)

	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var check model.Check
	if err := json.Unmarshal(data, &check); err != nil {
		return nil, err
	}

	return &check, nil
}
