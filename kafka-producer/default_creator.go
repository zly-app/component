package kafka_producer

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &kafkaProducer{
	connSync:  conn.NewConn(),
	connAsync: conn.NewConn(),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

// 获取kafka同步生产者
func GetClient(name string) Client {
	return defCreator.GetClient(name)
}

// 获取kafka同步生产者
func GetDefClient() Client {
	return defCreator.GetDefClient()
}

// 获取kafka异步生产者
func GetAsyncClient(name string) AsyncClient {
	return defCreator.GetAsyncClient(name)
}

// 获取kafka异步生产者
func GetDefAsyncClient() AsyncClient {
	return defCreator.GetDefAsyncClient()
}
