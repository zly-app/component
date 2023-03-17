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

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
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

func (r *RedisCreatorAdapter) makeClient(name string) (conn.IInstance, error) {
	conf := newRedisConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("创建redis客户端失败: %v", err)
	}
	return &instance{client: client}, nil
}

func (r *RedisCreatorAdapter) Close() {
	r.conn.CloseAll()
}
