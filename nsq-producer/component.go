/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package nsq_producer

import (
	"fmt"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
)

type INsqProducer interface {
	// 获取nsq发布者
	GetNsqProducer(name ...string) *nsq.Producer
	// 获取nsq发布者
	GetDefNsqProducer() *nsq.Producer
	// 关闭
	Close()
}

type instance struct {
	producer *nsq.Producer
}

func (i *instance) Close() {
	i.producer.Stop()
}

type NsqProducer struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewNsqProducer(app core.IApp, componentType ...core.ComponentType) INsqProducer {
	n := &NsqProducer{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		n.componentType = componentType[0]
	}
	return n
}

func (r *NsqProducer) GetNsqProducer(name ...string) *nsq.Producer {
	return r.conn.GetInstance(r.makeClient, name...).(*instance).producer
}

func (r *NsqProducer) GetDefNsqProducer() *nsq.Producer {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).producer
}

func (r *NsqProducer) makeClient(name string) (conn.IInstance, error) {
	conf := newConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err != nil {
		return nil, err
	}
	if err = conf.Check(); err != nil {
		return nil, fmt.Errorf("组件%s的配置错误: %s", r.componentType, err)
	}

	nsqConf := nsq.NewConfig()
	nsqConf.AuthSecret = conf.AuthSecret
	nsqConf.HeartbeatInterval = time.Duration(conf.HeartbeatInterval) * time.Millisecond
	nsqConf.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Millisecond
	nsqConf.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Millisecond
	nsqConf.DialTimeout = time.Duration(conf.DialTimeout) * time.Millisecond

	producer, err := nsq.NewProducer(conf.Address, nsqConf)
	return &instance{producer}, err
}

func (r *NsqProducer) Close() {
	r.conn.CloseAll()
}
