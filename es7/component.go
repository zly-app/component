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
	stdhttp "net/http"
	"strings"
	"time"

	elastic7 "github.com/olivere/elastic/v7"
	"github.com/zly-app/component/http"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
)

type Creator interface {
	// 获取es7客户端
	GetClient(name string) (*Client, error)
	// 获取默认es7客户端
	GetDefClient() (*Client, error)
}

type ES7 struct {
	conn *conn.AnyConn[*Client]
}

func GetCreator() Creator {
	return defCreator
}

func (e *ES7) GetClient(name string) (*Client, error) {
	client, err := e.conn.GetConn(e.makeClient, name)
	return client, err
}

func (e *ES7) GetDefClient() (*Client, error) {
	client, err := e.conn.GetConn(e.makeClient, consts.DefaultComponentName)
	return client, err
}

func (e *ES7) makeClient(name string) (*Client, error) {
	conf := newConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		return nil, fmt.Errorf("es7服务配置错误: %v", err)
	}

	opts := []elastic7.ClientOptionFunc{
		elastic7.SetURL(strings.Split(conf.Address, ",")...),
		elastic7.SetSniff(conf.Sniff),
		elastic7.SetHealthcheck(conf.HealthCheck),
		elastic7.SetGzip(conf.GZip),
		elastic7.SetHttpClient(&stdhttp.Client{Transport: http.NewTransport(name, conf.InsecureSkipVerify)}),
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
	return client, err
}

func (e *ES7) Close() {
	e.conn.CloseAll()
}
