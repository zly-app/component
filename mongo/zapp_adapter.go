package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "mongo"

type Creator interface {
	// 获取mongo客户端
	GetClient(name string) *Client
	// 获取默认mongo客户端
	GetDefClient() *Client
}

type instance struct {
	client         *Client
	connectTimeout time.Duration
}

func (i *instance) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), i.connectTimeout)
	defer cancel()
	_ = i.client.Disconnect(ctx)
}

type mongoCreator struct {
	conn *conn.Conn
}

func GetCreator() Creator {
	return defCreator
}

func (r *mongoCreator) GetClient(name string) *Client {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *mongoCreator) GetDefClient() *Client {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).client
}

func (r *mongoCreator) makeClient(name string) (conn.IInstance, error) {
	conf := NewMongoConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析mongo客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("创建mongo客户端失败: %v", err)
	}
	return &instance{client: client, connectTimeout: time.Duration(conf.ConnectTimeout) * time.Second}, nil
}

func (r *mongoCreator) Close() {
	r.conn.CloseAll()
}
