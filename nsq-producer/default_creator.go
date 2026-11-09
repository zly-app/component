package nsq_producer

import (
	"github.com/nsqio/go-nsq"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

// NewCreator 创建 nsq 发布者集合。实例由组合根持有并负责 Close，不再有包级单例。
func NewCreator(app core.IApp) Creator {
	return &nsqCreator{
		app: app,
		conn: conn.NewAnyConn[*nsq.Producer](func(name string, conn *nsq.Producer) {
			conn.Stop()
		}),
	}
}
