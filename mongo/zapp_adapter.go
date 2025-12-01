package mongo

import (
	"fmt"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "mongo"

type Creator interface {
	// 获取mongo客户端
	GetClient(name string) (*Client, error)
	// 获取默认mongo客户端
	GetDefClient() (*Client, error)
}

type mongoCreator struct {
	conn *conn.AnyConn[*Client]
}

func GetCreator() Creator {
	return defCreator
}

func (r *mongoCreator) GetClient(name string) (*Client, error) {
	return r.conn.GetConn(r.makeClient, name)
}

func (r *mongoCreator) GetDefClient() (*Client, error) {
	return r.conn.GetConn(r.makeClient, consts.DefaultComponentName)
}

func (r *mongoCreator) makeClient(name string) (*Client, error) {
	conf := NewMongoConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析mongo客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("创建mongo客户端失败: %v", err)
	}
	return client, nil
}

func (r *mongoCreator) Close() {
	r.conn.CloseAll()
}
