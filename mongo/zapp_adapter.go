package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "mongo"

// 当前组件类型
var nowComponentType = DefaultComponentType

// 设置组件类型, 这个函数应该在 zapp.NewApp 之前调用
func SetComponentType(t core.ComponentType) {
	nowComponentType = t
}

type IMongoCreator interface {
	// 获取mongo客户端
	GetMongo(name string) *Client
	// 获取默认mongo客户端
	GetDefMongo() *Client
	// 关闭
	Close()
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

type MongoCreatorAdapter struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewMongoCreator(app core.IApp) IMongoCreator {
	r := &MongoCreatorAdapter{
		app:           app,
		conn:          conn.NewConn(),
		componentType: nowComponentType,
	}
	return r
}

func (r *MongoCreatorAdapter) GetMongo(name string) *Client {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *MongoCreatorAdapter) GetDefMongo() *Client {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).client
}

func (r *MongoCreatorAdapter) makeClient(name string) (conn.IInstance, error) {
	conf := NewMongoConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析mongo客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("创建mongo客户端失败: %v", err)
	}
	return &instance{client: client, connectTimeout: time.Duration(conf.ConnectTimeout) * time.Second}, nil
}

func (r *MongoCreatorAdapter) Close() {
	r.conn.CloseAll()
}
