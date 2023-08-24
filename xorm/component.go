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
	"runtime"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	_ "github.com/denisenkom/go-mssqldb" // mssql
	_ "github.com/go-sql-driver/mysql"   // mysql
	_ "github.com/lib/pq"                // postgres
	_ "github.com/mattn/go-sqlite3"      // sqlite
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/pkg/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
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
	name          string

	tracer   trace.Tracer
	spanOpts []trace.SpanStartOption
}

type IXormCreator interface {
	// 获取
	GetXorm(name string) *Engine
	// 获取
	GetDefXorm() *Engine
	// 释放
	Close()
}

type instance struct {
	*Engine
}

func (i *instance) Close() {
	_ = i.Engine.Close()
}

// 创建xorm建造者
func NewXormCreator(app core.IApp, componentType ...core.ComponentType) IXormCreator {
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

func (x *Xorm) GetXorm(name string) *Engine {
	return x.conn.GetInstance(x.makeClient, name).(*instance).Engine
}

func (x *Xorm) GetDefXorm() *Engine {
	return x.conn.GetInstance(x.makeClient, consts.DefaultComponentName).(*instance).Engine
}

func (x *Xorm) makeClient(name string) (conn.IInstance, error) {
	x.name = name
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
	e.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetimeSec) * time.Second)

	e.SetTableMapper(x.makeNameMapper(conf.TableMapperRule))
	e.SetColumnMapper(x.makeNameMapper(conf.ColumnMapperRule))

	if l, err := time.LoadLocation(conf.TZ); err == nil {
		e.SetTZLocation(l)
	}

	if !conf.DisableOpenTrace {
		e.AddHook(x)

		const instrumName = "github.com/zly-app/component/xorm"
		x.tracer = otel.GetTracerProvider().Tracer(instrumName)
		x.spanOpts = []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(semconv.DBNameKey.String(name)),
		}
	}
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

func (x *Xorm) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
	fn, file, line := funcFileLine("xorm.io/xorm")

	attrs := make([]attribute.KeyValue, 0, 8)
	attrs = append(attrs,
		semconv.CodeFunctionKey.String(fn),
		semconv.CodeFilepathKey.String(file),
		semconv.CodeLineNumberKey.Int(line),
	)

	opts := append([]trace.SpanStartOption{}, x.spanOpts...)
	opts = append(opts, trace.WithAttributes(attrs...))

	ctx, _ := x.tracer.Start(c.Ctx, "xorm_sql."+x.name, opts...)

	args, _ := sonic.MarshalString(c.Args)
	utils.Otel.CtxEvent(ctx, "send",
		utils.OtelSpanKey("sql").String(c.SQL),
		utils.OtelSpanKey("args").String(args),
	)
	return ctx, nil
}

func (x *Xorm) AfterProcess(c *contexts.ContextHook) error {
	defer utils.Otel.CtxEnd(c.Ctx)

	if c.Err != nil {
		utils.Otel.CtxErrEvent(c.Ctx, "recv", c.Err)
	}
	return nil
}

func funcFileLine(pkg string) (string, string, int) {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])

	var fn, file string
	var line int
	for {
		f, ok := ff.Next()
		if !ok {
			break
		}
		fn, file, line = f.Function, f.File, f.Line
		if !strings.Contains(fn, pkg) {
			break
		}
	}

	if ind := strings.LastIndexByte(fn, '/'); ind != -1 {
		fn = fn[ind+1:]
	}

	return fn, file, line
}
