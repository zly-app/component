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

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "redis"

type IRedisCreator interface {
	// 获取redis客户端
	GetRedis(name string) UniversalClient
	// 获取默认redis客户端
	GetDefRedis() UniversalClient
}

type instance struct {
	client UniversalClient
}

func (i *instance) Close() {
	_ = i.client.Close()
}

type RedisCreatorAdapter struct {
	conn *conn.Conn
}

// deprecated: use GetRedisCreator
func NewRedisCreator(app core.IApp) IRedisCreator {
	return defCreator
}

func GetRedisCreator() IRedisCreator {
	return defCreator
}

func (r *RedisCreatorAdapter) GetRedis(name string) UniversalClient {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *RedisCreatorAdapter) GetDefRedis() UniversalClient {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).client
}

func (r *RedisCreatorAdapter) makeClient(name string) (conn.IInstance, error) {
	conf := NewRedisConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("创建redis客户端失败: %v", err)
	}

	if err := InstrumentTracing(string(DefaultComponentType), name, client); err != nil {
		logger.Log.Error("redisotel.InstrumentTracing err", err)
	}

	if err := redisotel.InstrumentMetrics(client); err != nil {
		logger.Log.Error("redisotel.InstrumentMetrics err", err)
	}
	return &instance{client: client}, nil
}

func (r *RedisCreatorAdapter) Close() {
	r.conn.CloseAll()
}
