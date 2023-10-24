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
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "redis"

// 当前组件类型
var nowComponentType = DefaultComponentType

// 设置组件类型, 这个函数应该在 zapp.NewApp 之前调用
func SetComponentType(t core.ComponentType) {
	nowComponentType = t
}

type IRedisCreator interface {
	// 获取redis客户端
	GetRedis(name string) UniversalClient
	// 获取默认redis客户端
	GetDefRedis() UniversalClient
	// 关闭
	Close()
}

type instance struct {
	client UniversalClient
}

func (i *instance) Close() {
	_ = i.client.Close()
}

type RedisCreatorAdapter struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewRedisCreator(app core.IApp) IRedisCreator {
	r := &RedisCreatorAdapter{
		app:           app,
		conn:          conn.NewConn(),
		componentType: nowComponentType,
	}
	return r
}

func (r *RedisCreatorAdapter) GetRedis(name string) UniversalClient {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *RedisCreatorAdapter) GetDefRedis() UniversalClient {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).client
}

func (r *RedisCreatorAdapter) makeClient(name string) (conn.IInstance, error) {
	conf := NewRedisConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("创建redis客户端失败: %v", err)
	}

	if err := InstrumentTracing(string(r.componentType), name, client); err != nil {
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
