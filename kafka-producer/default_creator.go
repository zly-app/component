package kafka_producer

import (
	"github.com/Shopify/sarama"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

// NewCreator 创建 kafka 生产者集合。实例由组合根持有并负责 Close，不再有包级单例。
func NewCreator(app core.IApp) Creator {
	return &kafkaProducer{
		app: app,
		connSync: conn.NewAnyConn[Client](func(name string, conn Client) {
			v, ok := conn.(sarama.SyncProducer)
			if ok {
				_ = v.Close()
			}
		}),
		connAsync: conn.NewAnyConn[AsyncClient](func(name string, conn AsyncClient) {
			v, ok := conn.(sarama.AsyncProducer)
			if ok {
				_ = v.Close()
			}
		}),
	}
}
