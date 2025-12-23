package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store interface {
	Get(ctx context.Context, key string, v any) error
	Set(ctx context.Context, key string, v any, expireAt time.Time) error
	SetTTL(ctx context.Context, key string, v any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type Redis struct {
	client *redis.Client
	prefix string
}

func NewRedis(addr, password, prefix string) (*Redis, error) {
	var opts *redis.Options

	// unix socket detection
	if strings.HasPrefix(addr, "/") {
		opts = &redis.Options{
			Network:  "unix",
			Addr:     addr,
			Password: password,
		}
	} else {
		opts = &redis.Options{
			Addr:     addr,
			Password: password,
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{
		client: client,
		prefix: prefix,
	}, nil
}

func (r *Redis) key(k string) string {
	return r.prefix + "-" + k
}

func (r *Redis) Get(ctx context.Context, key string, v any) error {
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (r *Redis) Set(ctx context.Context, key string, v any, expireAt time.Time) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(key), data, time.Until(expireAt)).Err()
}

func (r *Redis) SetTTL(ctx context.Context, key string, v any, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(key), data, ttl).Err()
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.key(key)).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, r.key(key)).Result()
	return n > 0, err
}

func (r *Redis) Close() error {
	return r.client.Close()
}
