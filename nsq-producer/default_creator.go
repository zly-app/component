package nsq_producer

import (
	"github.com/nsqio/go-nsq"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &nsqCreator{
	conn: conn.NewAnyConn[*nsq.Producer](func(name string, conn *nsq.Producer) {
		conn.Stop()
	}),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

// 获取客户端
func GetClient(name string) (*nsq.Producer, error) {
	return defCreator.GetClient(name)
}

// 获取默认客户端
func GetDefClient() (*nsq.Producer, error) {
	return defCreator.GetDefClient()
}
