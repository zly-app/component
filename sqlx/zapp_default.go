package sqlx

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var defCreator = &SqlxCreator{
	conn: conn.NewConn(),
}

func init() {
	zapp.AddHandler(zapp.AfterCloseComponent, func(_ core.IApp, _ handler.HandlerType) {
		defCreator.Close()
	})
}

func GetClient(name string) Client {
	return defCreator.GetClient(name)
}
func GetDefClient() Client {
	return defCreator.GetDefClient()
}
