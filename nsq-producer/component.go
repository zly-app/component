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
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
)

type Creator interface {
	// 获取nsq发布者
	GetClient(name string) *nsq.Producer
	// 获取nsq发布者
	GetDefClient() *nsq.Producer
}

type instance struct {
	producer *nsq.Producer
}

func (i *instance) Close() {
	i.producer.Stop()
}

type nsqCreator struct {
	conn *conn.Conn
}

func GetCreator() Creator {
	return defCreator
}

func (r *nsqCreator) GetClient(name string) *nsq.Producer {
	return r.conn.GetInstance(r.makeClient, name).(*instance).producer
}

func (r *nsqCreator) GetDefClient() *nsq.Producer {
	return r.conn.GetInstance(r.makeClient, consts.DefaultComponentName).(*instance).producer
}

func (r *nsqCreator) makeClient(name string) (conn.IInstance, error) {
	conf := newConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, err
	}
	if err = conf.Check(); err != nil {
		return nil, fmt.Errorf("组件%s的配置错误: %s", DefaultComponentType, err)
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

func (r *nsqCreator) Close() {
	r.conn.CloseAll()
}
