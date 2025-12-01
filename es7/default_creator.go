package es7

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &ES7{
	conn: conn.NewAnyConn[*Client](func(name string, conn *Client) {
		conn.Stop()
	}),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

// 获取户端
func GetClient(name string) (*Client, error) {
	return defCreator.GetClient(name)
}

// 获取默认客户端
func GetDefClient() (*Client, error) {
	return defCreator.GetDefClient()
}
