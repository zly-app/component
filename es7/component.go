/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package es7

import (
	"context"
	"fmt"
	"strings"
	"time"

	elastic7 "github.com/olivere/elastic/v7"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type IES7Component interface {
	// 获取es7客户端
	GetES7(name ...string) *elastic7.Client
	// 关闭
	Close()
}

type instance struct {
	*elastic7.Client
}

func (i *instance) Close() {
	i.Client.Stop()
}

type ES7 struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

// 创建es7组件
func NewES7(app core.IApp, componentType ...core.ComponentType) IES7Component {
	e := &ES7{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		e.componentType = componentType[0]
	}
	return e
}

func (e *ES7) GetES7(name ...string) *elastic7.Client {
	return e.conn.GetInstance(e.makeClient, name...).(*instance).Client
}

func (e *ES7) makeClient(name string) (conn.IInstance, error) {
	var conf ES7Config
	err := e.app.GetConfig().ParseComponentConfig(e.componentType, name, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Address == "" {
		return nil, fmt.Errorf("%s的address为空", e.componentType)
	}

	opts := []elastic7.ClientOptionFunc{
		elastic7.SetURL(strings.Split(conf.Address, ",")...),
		elastic7.SetSniff(conf.Sniff),
		elastic7.SetHealthcheck(conf.Healthcheck == nil || *conf.Healthcheck),
		elastic7.SetGzip(conf.GZip),
	}
	if conf.UserName != "" || conf.Password != "" {
		opts = append(opts, elastic7.SetBasicAuth(conf.UserName, conf.Password))
	}
	if conf.Retry > 0 {
		ticks := make([]int, conf.Retry)
		for i := 0; i < conf.Retry; i++ {
			ticks[i] = conf.RetryInterval
		}
		opts = append(opts, elastic7.SetRetrier(elastic7.NewBackoffRetrier(elastic7.NewSimpleBackoff(ticks...))))
	}

	ctx := context.Background()
	if conf.DialTimeout > 0 {
		c, cancel := context.WithTimeout(ctx, time.Duration(conf.DialTimeout)*time.Millisecond)
		defer cancel()
		ctx = c
	}

	client, err := elastic7.DialContext(ctx, opts...)
	return &instance{client}, err
}

func (e *ES7) Close() {
	e.conn.CloseAll()
}
