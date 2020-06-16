package redis

import "gopkg.in/redis.v5"

// Client Represents the redis client
type Client struct {
	*redis.Client
}
