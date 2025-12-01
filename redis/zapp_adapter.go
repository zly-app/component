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
	GetClient(name string) (UniversalClient, error)
	// 获取默认redis客户端
	GetDefClient() (UniversalClient, error)
}

type RedisCreatorAdapter struct {
	conn *conn.AnyConn[UniversalClient]
}

func GetCreator() Creator {
	return defCreator
}

func (r *RedisCreatorAdapter) GetClient(name string) (UniversalClient, error) {
	return r.conn.GetConn(r.makeClient, name)
}

func (r *RedisCreatorAdapter) GetDefClient() (UniversalClient, error) {
	return r.conn.GetConn(r.makeClient, consts.DefaultComponentName)
}

func (r *RedisCreatorAdapter) makeClient(name string) (UniversalClient, error) {
	conf := NewRedisConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf, name)
	if err != nil {
		return nil, fmt.Errorf("创建redis客户端失败: %v", err)
	}
	return client, nil
}

func (r *RedisCreatorAdapter) Close() {
	r.conn.CloseAll()
}
