package pulsar_producer

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &producerCreator{
	conn: conn.NewAnyConn[Client](func(name string, conn Client) {
		v, ok := conn.(*PulsarProducer)
		if ok {
			v.Close()
		}
	}),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

// 获取客户端
func GetClient(name string) Client {
	return defCreator.GetClient(name)
}

// 获取默认客户端
func GetDefClient() Client {
	return defCreator.GetDefClient()
}
