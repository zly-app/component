package redis

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &RedisCreatorAdapter{
	conn: conn.NewAnyConn[UniversalClient](func(name string, conn UniversalClient) {
		_ = conn.Close()
	}),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

// 获取redis客户端
func GetClient(name string) (UniversalClient, error) {
	return defCreator.GetClient(name)
}

// 获取默认redis客户端
func GetDefClient() (UniversalClient, error) {
	return defCreator.GetDefClient()
}
