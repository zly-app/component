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

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "redis"

type Creator interface {
	// 获取redis客户端
	GetClient(name string) UniversalClient
	// 获取默认redis客户端
	GetDefClient() UniversalClient
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

func GetCreator() Creator {
	return defCreator
}

func (r *RedisCreatorAdapter) GetClient(name string) UniversalClient {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *RedisCreatorAdapter) GetDefClient() UniversalClient {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).client
}

func (r *RedisCreatorAdapter) makeClient(name string) (conn.IInstance, error) {
	conf := NewRedisConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf, name)
	if err != nil {
		return nil, fmt.Errorf("创建redis客户端失败: %v", err)
	}
	return &instance{client: client}, nil
}

func (r *RedisCreatorAdapter) Close() {
	r.conn.CloseAll()
}
