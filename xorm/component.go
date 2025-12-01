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
	"github.com/zly-app/zapp"
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
)

type xormCreator struct {
	conn *conn.AnyConn[*Engine]
	name string

	tracer   trace.Tracer
	spanOpts []trace.SpanStartOption
}

type Creator interface {
	// 获取
	GetClient(name string) (*Engine, error)
	// 获取
	GetDefClient() (*Engine, error)
}

func GetCreator() Creator {
	return defCreator
}

func (x *xormCreator) GetClient(name string) (*Engine, error) {
	return x.conn.GetConn(x.makeClient, name)
}

func (x *xormCreator) GetDefClient() (*Engine, error) {
	return x.conn.GetConn(x.makeClient, consts.DefaultComponentName)
}

func (x *xormCreator) makeClient(name string) (*Engine, error) {
	x.name = name
	conf := newConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
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

	e.AddHook(x)

	const instrumName = "github.com/zly-app/component/xorm"
	x.tracer = otel.GetTracerProvider().Tracer(instrumName)
	x.spanOpts = []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.DBNameKey.String(name)),
	}

	return e, nil
}

func (x *xormCreator) makeNameMapper(rule string) names.Mapper {
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

func (x *xormCreator) Close() {
	x.conn.CloseAll()
}

// -------------
//  xorm  hook
// -------------

func (x *xormCreator) BeforeProcess(c *contexts.ContextHook) (context.Context, error) {
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
	utils.Trace.CtxEvent(ctx, "send",
		utils.OtelSpanKey("sql").String(c.SQL),
		utils.OtelSpanKey("args").String(args),
	)
	return ctx, nil
}

func (x *xormCreator) AfterProcess(c *contexts.ContextHook) error {
	defer utils.Trace.CtxEnd(c.Ctx)

	if c.Err != nil {
		utils.Trace.CtxErrEvent(c.Ctx, "recv", c.Err)
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
