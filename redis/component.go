/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package redis

import (
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

type IRedisCreator interface {
	// 获取redis客户端
	GetRedis(name string) UniversalClient
	// 关闭
	Close()
}

type instance struct {
	client UniversalClient
}

func (i *instance) Close() {
	_ = i.client.Close()
}

type Redis struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewRedisCreator(app core.IApp, componentType ...core.ComponentType) IRedisCreator {
	r := &Redis{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		r.componentType = componentType[0]
	}
	return r
}

func (r *Redis) GetRedis(name string) UniversalClient {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *Redis) makeClient(name string) (conn.IInstance, error) {
	conf := newRedisConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("生成redis客户端失败: %v", err)
	}
	return &instance{client: client}, nil
}

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

func (r *Redis) Close() {
	r.conn.CloseAll()
}
