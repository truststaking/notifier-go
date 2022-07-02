package redis

import (
	"context"

	"github.com/ElrondNetwork/notifier-go/config"
	"github.com/go-redis/redis/v8"
)

// CreateSimpleClient will create a redis client for a redis setup with one instance
func CreateSimpleClient(cfg config.RedisConfig) (RedLockClient, error) {
	opt, err := redis.ParseURL(cfg.Url)
	if err != nil {
		return nil, err
	}
	// opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	// opt.Addr = "trustmarket.redis.cache.windows.net:6380"
	// opt.Password = "sD2VfqU37NdUl7AoA4H9juXt9sFw6YhDNAzCaJr3vbA="
	client := redis.NewClient(opt)

	rc := NewRedisClientWrapper(client)
	ok := rc.IsConnected(context.Background())
	if !ok {
		return nil, ErrRedisConnectionFailed
	}

	return rc, nil
}

// CreateFailoverClient will create a redis client for a redis setup with sentinel
func CreateFailoverClient(cfg config.RedisConfig) (RedLockClient, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: []string{cfg.SentinelUrl},
	})
	rc := NewRedisClientWrapper(client)

	ok := rc.IsConnected(context.Background())
	if !ok {
		return nil, ErrRedisConnectionFailed
	}

	return rc, nil
}
