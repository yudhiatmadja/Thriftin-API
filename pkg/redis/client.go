package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func New(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil { return nil, err }
	c := redis.NewClient(opt)
	if err := c.Ping(context.Background()).Err(); err != nil { return nil, err }
	return c, nil
}
