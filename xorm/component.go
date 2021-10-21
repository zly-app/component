/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/21
   Description :
-------------------------------------------------
*/

package xorm

import (
	"context"
	"fmt"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb" // mssql
	_ "github.com/go-sql-driver/mysql"   // mysql
	_ "github.com/lib/pq"                // postgres
	_ "github.com/mattn/go-sqlite3"      // sqlite
	"github.com/opentracing/opentracing-go"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"
	"xorm.io/xorm/names"

	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type Xorm struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

type IXorm interface {
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
func NewXorm(app core.IApp, componentType ...core.ComponentType) IXorm {
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
	conf := newConfig()
	err := x.app.GetConfig().ParseComponentConfig(x.componentType, name, conf)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		return nil, fmt.Errorf("xorm的配置错误: %v", err)
	}

	e, err := xorm.NewEngine(conf.Driver, conf.Source)
	if err != nil {
		return nil, err
	}
	e.SetMaxIdleConns(conf.MaxIdleConns)
	e.SetMaxOpenConns(conf.MaxOpenConns)
	e.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Millisecond)

	e.SetTableMapper(x.makeNameMapper(conf.TableMapperRule))
	e.SetColumnMapper(x.makeNameMapper(conf.ColumnMapperRule))

	e.AddHook(x)
	return &instance{e}, nil
}

func (x *Xorm) makeNameMapper(rule string) names.Mapper {
	switch strings.ToLower(rule) {
	case "SnakeMapper":
		return names.SnakeMapper{}
	case "SameMapper":
		return names.SameMapper{}
	case "GonicMapper":
		return names.GonicMapper{}
	}
	return names.GonicMapper{}
}

func (x *Xorm) Close() {
	x.conn.CloseAll()
}

// -------------
//  xorm  hook
// -------------

type contextKey struct{}

var xormSpanKey = &contextKey{}

func (x *Xorm) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	var span opentracing.Span
	parentSpan := opentracing.SpanFromContext(c.Ctx) // 获取父span
	if parentSpan != nil {
		span = opentracing.StartSpan("xorm_sql", opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan("xorm_sql")
	}

	// 存入上下文
	c.Ctx = context.WithValue(c.Ctx, xormSpanKey, span)
	return c.Ctx, nil
}

func (x *Xorm) AfterProcess(c *contexts.ContextHook) error {
	span, ok := c.Ctx.Value(xormSpanKey).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	if c.Err != nil {
		span.SetTag("error", true)
		span.SetTag("err", c.Err.Error())
	}

	span.SetTag("exec_time", c.ExecuteTime.String())
	span.SetTag("sql", c.SQL)
	span.SetTag("args", c.Args)
	return nil
}
