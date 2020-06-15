package redis

import (
	"fmt"
	"sync"

	"github.com/SayedAlesawy/Videra-Storage/config"
	"gopkg.in/redis.v5"
)

// redisDriverOnce Used to garauntee thread safety for singleton instances
var redisDriverOnce sync.Once

// redisDriverInstance A singleton instance of the redis client object
var redisDriverInstance *Client

// Instance A function to obtain a new redis instance
func Instance(redisConfig *config.Redisconfig) (*Client, error) {
	var err error

	redisDriverOnce.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisConfig.Host, redisConfig.Port),
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})

		_, err = client.Ping().Result()

		redisDriverInstance = &Client{client}
	})

	return redisDriverInstance, err
}
