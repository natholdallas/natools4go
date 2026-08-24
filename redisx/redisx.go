// Package redisx is a tiny package for Redis client
package redisx

import "github.com/redis/go-redis/v9"

func New(opt *redis.Options) *redis.Client {
	if opt.Addr == "" {
		return nil
	}
	return redis.NewClient(opt)
}
