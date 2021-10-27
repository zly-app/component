/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type IRedis interface {
	// 获取redis客户端
	GetRedis(name ...string) redis.UniversalClient
	// 关闭
	Close()
}

type instance struct {
	client redis.UniversalClient
}

func (i *instance) Close() {
	_ = i.client.Close()
}

type Redis struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewRedis(app core.IApp, componentType ...core.ComponentType) IRedis {
	r := &Redis{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		r.componentType = componentType[0]
	}
	return r
}

func (r *Redis) GetRedis(name ...string) redis.UniversalClient {
	return r.conn.GetInstance(r.makeClient, name...).(*instance).client
}

func (r *Redis) makeClient(name string) (conn.IInstance, error) {
	conf := newRedisConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		return nil, fmt.Errorf("redis配置错误: %v", err)
	}

	var client redis.UniversalClient
	if conf.IsCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(conf.Address, ","),
			Username:     conf.UserName,
			Password:     conf.Password,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(conf.WriteTimeout) * time.Millisecond,
			DialTimeout:  time.Duration(conf.DialTimeout) * time.Millisecond,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         conf.Address,
			Username:     conf.UserName,
			Password:     conf.Password,
			DB:           conf.DB,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(conf.WriteTimeout) * time.Millisecond,
			DialTimeout:  time.Duration(conf.DialTimeout) * time.Millisecond,
		})
	}

	if conf.EnableOpenTrace {
		client.AddHook(r)
	}

	return &instance{client}, nil
}

func (r *Redis) Close() {
	r.conn.CloseAll()
}

// -------------
//  redis  hook
// -------------

type contextKey struct{}

var redisSpanKey = &contextKey{}

func (r *Redis) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	var span opentracing.Span
	parentSpan := opentracing.SpanFromContext(ctx) // 获取父span
	if parentSpan != nil {
		span = opentracing.StartSpan("redis_cmds", opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan("redis_cmds")
	}

	return context.WithValue(ctx, redisSpanKey, span), nil
}

func (r *Redis) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span, ok := ctx.Value(redisSpanKey).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		span.SetTag("error", true)
		span.SetTag("err", cmd.Err())
	}
	span.SetTag("cmd", cmd.FullName())
	span.SetTag("args", cmd.Args())
	return nil
}

func (r *Redis) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	var span opentracing.Span
	parentSpan := opentracing.SpanFromContext(ctx) // 获取父span
	if parentSpan != nil {
		span = opentracing.StartSpan("redis_multi_cmds", opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan("redis_multi_cmds")
	}

	return context.WithValue(ctx, redisSpanKey, span), nil
}

func (r *Redis) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span, ok := ctx.Value(redisSpanKey).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	var err error
	for _, cmd := range cmds {
		sp := opentracing.StartSpan(cmd.FullName(), opentracing.ChildOf(span.Context()))

		if cmd.Err() != nil && cmd.Err() != redis.Nil {
			if err == nil {
				err = cmd.Err()
			}

			sp.SetTag("error", true)
			sp.SetTag("err", cmd.Err())
		}
		sp.SetTag("cmd", cmd.FullName())
		sp.SetTag("args", cmd.Args())
		sp.Finish()
	}

	if err != nil {
		span.SetTag("error", true)
		span.SetTag("err", err.Error())
	}

	span.SetTag("cmd count", len(cmds))
	return nil
}
