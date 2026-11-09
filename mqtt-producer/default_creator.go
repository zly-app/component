package mqtt_producer

import (
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

// NewCreator 创建 mqtt 生产者集合。实例由组合根持有并负责 Close，不再有包级单例。
func NewCreator(app core.IApp) Creator {
	return &mqttProducerCreator{
		app: app,
		conn: conn.NewAnyConn[Client](func(name string, conn Client) {
			v, ok := conn.(*MQTTProducer)
			if ok {
				v.Close()
			}
		}),
	}
}
