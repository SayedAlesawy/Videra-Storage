package config

import (
	"sync"
)

// Redisconfig Houses the configurations of the redis instance
type Redisconfig struct {
	Host     string //Redis host
	Port     string //Redis port
	Password string //Redis DB password
	DB       int    //Redis DB number
	PoolSize int    //Redis pool size
}

// redisConfigOnce Used to garauntee thread safety for singleton instances
var redisConfigOnce sync.Once

// redisConfigInstance A singleton instance of the redis config object
var redisConfigInstance *Redisconfig

// RedisConfig A function to read redis config
func (manager *ConfigurationManager) RedisConfig() *Redisconfig {
	redisConfigOnce.Do(func() {
		redisConfig := Redisconfig{
			Host:     envString("REDIS_HOST", "127.0.0.1"),
			Port:     envString("REDIS_PORT", "6379"),
			Password: envString("REDIS_PASSWORD", ""),
			DB:       int(envInt("REDIS_DB", "0")),
			PoolSize: int(envInt("REDIS_POOL_SIZE", "10")),
		}

		redisConfigInstance = &redisConfig
	})

	return redisConfigInstance
}
