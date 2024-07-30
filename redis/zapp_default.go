package redis

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &RedisCreatorAdapter{
	conn: conn.NewConn(),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

// 获取redis客户端
func GetClient(name string) UniversalClient {
	return defCreator.GetRedis(name)
}

// 获取默认redis客户端
func GetDefClient() UniversalClient {
	return defCreator.GetDefRedis()
}
