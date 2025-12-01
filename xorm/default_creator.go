package xorm

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &xormCreator{
	conn: conn.NewAnyConn[*Engine](func(name string, conn *Engine) {
		_ = conn.Close()
	}),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

func GetClient(name string) (*Engine, error) {
	return defCreator.GetClient(name)
}
func GetDefClient() (*Engine, error) {
	return defCreator.GetDefClient()
}
