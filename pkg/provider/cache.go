package provider

import (
	"fmt"

	"github.com/0xsj/hexagonal-go/pkg/cache"
	"github.com/0xsj/hexagonal-go/pkg/cache/redis"
	goredis "github.com/redis/go-redis/v9"
)

// ProvideCache creates a cache instance.
// If caching is disabled, returns nil.
func ProvideCache(config cache.Config) (cache.Cache, error) {
	if !config.Enabled {
		return nil, nil
	}

	options := &goredis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return redis.New(options)
}
