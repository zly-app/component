package es7

import (
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

// NewCreator 创建 es7 客户端集合。实例由组合根持有并负责 Close，不再有包级单例。
func NewCreator(app core.IApp) Creator {
	return &ES7{
		app: app,
		conn: conn.NewAnyConn[*Client](func(name string, conn *Client) {
			conn.Stop()
		}),
	}
}
