package redis

import (
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(conf *RedisConfig) (UniversalClient, error) {
	err := conf.Check()
	if err != nil {
		return nil, fmt.Errorf("redis客户端配置错误: %v", err)
	}

	var client redis.UniversalClient
	client = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:                 strings.Split(conf.Address, ","),
		DB:                    conf.DB,
		Username:              conf.UserName,
		Password:              conf.Password,
		MaxRetries:            conf.MaxRetries,
		DialTimeout:           time.Duration(conf.ConnectTimeout) * time.Second,
		ReadTimeout:           time.Duration(conf.ReadTimeoutSec) * time.Second,
		WriteTimeout:          time.Duration(conf.WriteTimeoutSec) * time.Second,
		ContextTimeoutEnabled: false, // 如果设为false, 传入的ctx无效, 如果设为true, 当出现超时时会导致client关闭(因为它发送或接收了一半的命令导致client不稳定)
		PoolSize:              conf.PoolSize,
		MinIdleConns:          conf.MinIdle,
		MaxIdleConns:          conf.MaxIdle,
		ConnMaxIdleTime:       time.Duration(conf.IdleTimeout) * time.Second,
		ConnMaxLifetime:       time.Duration(conf.MaxConnLifetime) * time.Second,
		MaxRedirects:          conf.MaxRetries,
	})
	return client, nil
}
