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

	"github.com/go-redis/redis/v8"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type IRedis interface {
	// 获取redis客户端
	GetRedis(name ...string) redis.UniversalClient
	// 关闭
	Close()
}

type instance struct {
	client redis.UniversalClient
}

func (i *instance) Close() {
	_ = i.client.Close()
}

type Redis struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewRedis(app core.IApp, componentType ...core.ComponentType) IRedis {
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

func (r *Redis) GetRedis(name ...string) redis.UniversalClient {
	return r.conn.GetInstance(r.makeClient, name...).(*instance).client
}

func (r *Redis) makeClient(name string) (conn.IInstance, error) {
	var conf RedisConfig
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Address == "" {
		return nil, fmt.Errorf("%s的address为空", r.componentType)
	}

	var client redis.UniversalClient
	if conf.IsCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(conf.Address, ","),
			Username:     conf.UserName,
			Password:     conf.Password,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(conf.WriteTimeout) * time.Millisecond,
			DialTimeout:  time.Duration(conf.DialTimeout) * time.Millisecond,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         conf.Address,
			Username:     conf.UserName,
			Password:     conf.Password,
			DB:           conf.DB,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(conf.WriteTimeout) * time.Millisecond,
			DialTimeout:  time.Duration(conf.DialTimeout) * time.Millisecond,
		})
	}

	return &instance{client}, nil
}

func (r *Redis) Close() {
	r.conn.CloseAll()
}
