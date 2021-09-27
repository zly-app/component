/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/1
   Description :
-------------------------------------------------
*/

package ssdb

import (
	"fmt"
	"net"

	"github.com/seefan/gossdb/v2"
	rconf "github.com/seefan/gossdb/v2/conf"
	"github.com/seefan/gossdb/v2/pool"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type ISsdb interface {
	// 获取客户端, 使用完后记得关闭
	GetSsdb(name ...string) *pool.Client
	// 关闭
	Close()
}

type instance struct {
	*pool.Connectors
}

func (i *instance) Close() {
	i.Connectors.Close()
}

type Ssdb struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

// 创建ssdb组件
func NewSsdb(app core.IApp, componentType ...core.ComponentType) ISsdb {
	s := &Ssdb{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		s.componentType = componentType[0]
	}
	return s
}

func (s *Ssdb) GetSsdb(name ...string) *pool.Client {
	return s.conn.GetInstance(s.makeClient, name...).(*instance).Connectors.GetClient()
}

func (s *Ssdb) makeClient(name string) (conn.IInstance, error) {
	conf := newSsdbConfig()
	err := s.app.GetConfig().ParseComponentConfig(s.componentType, name, conf)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		return nil, fmt.Errorf("ssdb配置错误: %v", err)
	}

	addr, err := net.ResolveTCPAddr("tcp", conf.Address)
	if err != nil {
		return nil, fmt.Errorf("ssdb配置错误, 无法解析addres: %v", err)
	}

	p, err := gossdb.NewPool(&rconf.Config{
		Host:             addr.IP.String(),
		Port:             addr.Port,
		Password:         conf.Password,
		GetClientTimeout: conf.GetClientTimeout / 1e3,
		MaxWaitSize:      conf.GetClientWaitQueueSize,
		ReadTimeout:      conf.ReadTimeout / 1e3,
		WriteTimeout:     conf.WriteTimeout / 1e3,
		ConnectTimeout:   conf.DialTimeout / 1e3,
		MinPoolSize:      conf.MinIdleConns,
		MaxPoolSize:      conf.PoolSize,
		RetryEnabled:     conf.RetryEnabled,
	})

	return &instance{p}, err
}

func (s *Ssdb) Close() {
	s.conn.CloseAll()
}
