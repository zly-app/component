/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/21
   Description :
-------------------------------------------------
*/

package xorm

import (
	"time"

	_ "github.com/denisenkom/go-mssqldb" // mssql
	_ "github.com/go-sql-driver/mysql"   // mysql
	_ "github.com/lib/pq"                // postgres
	_ "github.com/mattn/go-sqlite3"      // sqlite
	"xorm.io/xorm"

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type Xorm struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

type IXormComponent interface {
	// 获取
	GetXorm(name ...string) *xorm.Engine
	// 释放
	Close()
}

type instance struct {
	*xorm.Engine
}

func (i *instance) Close() {
	_ = i.Engine.Close()
}

// 创建xorm组件
func NewXorm(app core.IApp, componentType ...core.ComponentType) IXormComponent {
	x := &Xorm{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		x.componentType = componentType[0]
	}
	return x
}

func (x *Xorm) GetXorm(name ...string) *xorm.Engine {
	return x.conn.GetInstance(x.makeClient, name...).(*instance).Engine
}

func (x *Xorm) makeClient(name string) (conn.IInstance, error) {
	var conf XormConfig
	err := x.app.GetConfig().ParseComponentConfig(x.componentType, name, &conf)
	if err != nil {
		return nil, err
	}

	e, err := xorm.NewEngine(conf.Driver, conf.Source)
	if err != nil {
		return nil, err
	}
	e.SetMaxIdleConns(conf.MaxIdleConns)
	e.SetMaxOpenConns(conf.MaxOpenConns)
	e.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Millisecond)
	return &instance{e}, nil
}

func (x *Xorm) Close() {
	x.conn.CloseAll()
}
