package redis

import (
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zly-app/zapp/logger"
)

func NewClient(conf *RedisConfig) (UniversalClient, error) {
	err := conf.Check()
	if err != nil {
		return nil, fmt.Errorf("redis客户端配置错误: %v", err)
	}

	var client redis.UniversalClient
	if conf.IsCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(conf.Address, ","),
			Username:     conf.UserName,
			Password:     conf.Password,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(conf.WriteTimeoutSec) * time.Second,
			DialTimeout:  time.Duration(conf.DialTimeoutSec) * time.Second,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         conf.Address,
			Username:     conf.UserName,
			Password:     conf.Password,
			DB:           conf.DB,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(conf.WriteTimeoutSec) * time.Second,
			DialTimeout:  time.Duration(conf.DialTimeoutSec) * time.Second,
		})
	}

	if err := redisotel.InstrumentTracing(client); err != nil {
		logger.Log.Error("redisotel.InstrumentTracing err", err)
	}

	if err := redisotel.InstrumentMetrics(client); err != nil {
		logger.Log.Error("redisotel.InstrumentMetrics err", err)
	}
	return client, nil
}
