package redis

import (
	"fmt"

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

// DefaultComponentType 默认组件类型
const DefaultComponentType core.ComponentType = "redis"

// Creator 是 redis 客户端集合：拥有「按名字复用的连接」这份数据，
// 以及「从组件配置建连」这条规则。实例由组合根创建并持有，不再有包级单例。
type Creator interface {
	// GetClient 获取指定名字的客户端
	GetClient(name string) (UniversalClient, error)
	// GetDefClient 获取默认客户端
	GetDefClient() (UniversalClient, error)
	// Close 关闭全部连接
	Close()
}

type creator struct {
	app  core.IApp
	conn *conn.AnyConn[UniversalClient]
}

// NewCreator 创建 redis 客户端集合。
func NewCreator(app core.IApp) Creator {
	return &creator{
		app: app,
		conn: conn.NewAnyConn[UniversalClient](func(_ string, client UniversalClient) {
			_ = client.Close()
		}),
	}
}

func (c *creator) GetClient(name string) (UniversalClient, error) {
	return c.conn.GetConn(c.makeClient, name)
}

func (c *creator) GetDefClient() (UniversalClient, error) {
	return c.conn.GetConn(c.makeClient, consts.DefaultComponentName)
}

func (c *creator) makeClient(name string) (UniversalClient, error) {
	conf := NewRedisConfig()
	if err := c.app.GetConfig().ParseComponentConfig(DefaultComponentType, name, conf); err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf, name)
	if err != nil {
		return nil, fmt.Errorf("创建redis客户端失败: %v", err)
	}
	InstrumentClient(name, client)
	return client, nil
}

func (c *creator) Close() {
	c.conn.CloseAll()
}
