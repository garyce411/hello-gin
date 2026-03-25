// 增加连接到redis的配置和连接池
package config

import (
	"context"
	redis "github.com/redis/go-redis/v9"
	"time"
)

// 使用单例模式
var redisClient *redis.Client

func NewRedisClient(address string, password string, db int) *redis.Client {
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
		Addr:           address,
		Password:       password,
		DB:             db,
		PoolSize:     10,
		MinIdleConns: 1,
		MaxIdleConns: 10,
		ConnMaxIdleTime: 10 * time.Minute,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			return nil
		},
	})}
	return redisClient
}
