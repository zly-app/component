package xorm

import (
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

// NewCreator 创建 xorm 引擎集合。实例由组合根持有并负责 Close，不再有包级单例。
func NewCreator(app core.IApp) Creator {
	return &xormCreator{
		app: app,
		conn: conn.NewAnyConn[*Engine](func(name string, conn *Engine) {
			_ = conn.Close()
		}),
	}
}
