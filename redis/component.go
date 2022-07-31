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
	open_log "github.com/opentracing/opentracing-go/log"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

type IRedisCreator interface {
	// 获取redis客户端
	GetRedis(name string) UniversalClient
	// 关闭
	Close()
}

type instance struct {
	client UniversalClient
}

func (i *instance) Close() {
	_ = i.client.Close()
}

type Redis struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType
}

func NewRedisCreator(app core.IApp, componentType ...core.ComponentType) IRedisCreator {
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

func (r *Redis) GetRedis(name string) UniversalClient {
	return r.conn.GetInstance(r.makeClient, name).(*instance).client
}

func (r *Redis) makeClient(name string) (conn.IInstance, error) {
	conf := newRedisConfig()
	err := r.app.GetConfig().ParseComponentConfig(r.componentType, name, conf)
	if err != nil {
		return nil, fmt.Errorf("解析redis客户端配置错误: %v", err)
	}

	client, err := NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("生成redis客户端失败: %v", err)
	}
	return &instance{client: client}, nil
}

func NewClient(conf *RedisConfig) (UniversalClient, error) {
	err := conf.Check()
	if err != nil {
		return nil, fmt.Errorf("redis客户端配置错误: %v", err)
	}

	var client redis.UniversalClient
	if conf.IsCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(conf.Address, ","),
			Username:     conf.UserName,
			Password:     conf.Password,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(conf.WriteTimeoutSec) * time.Second,
			DialTimeout:  time.Duration(conf.DialTimeoutSec) * time.Second,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         conf.Address,
			Username:     conf.UserName,
			Password:     conf.Password,
			DB:           conf.DB,
			MinIdleConns: conf.MinIdleConns,
			PoolSize:     conf.PoolSize,
			ReadTimeout:  time.Duration(conf.ReadTimeoutSec) * time.Second,
			WriteTimeout: time.Duration(conf.WriteTimeoutSec) * time.Second,
			DialTimeout:  time.Duration(conf.DialTimeoutSec) * time.Second,
		})
	}

	if !conf.DisableOpenTrace {
		client.AddHook(redisTraceHook{})
	}

	return client, nil
}

func (r *Redis) Close() {
	r.conn.CloseAll()
}

// -------------
//  redis  hook
// -------------

type contextKey struct{}

var redisSpanKey = &contextKey{}

type redisTraceHook struct{}

func (redisTraceHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	var span opentracing.Span
	parentSpan := opentracing.SpanFromContext(ctx) // 获取父span
	if parentSpan != nil {
		span = opentracing.StartSpan("redis", opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan("redis")
	}

	return context.WithValue(ctx, redisSpanKey, span), nil
}

func (redisTraceHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span, ok := ctx.Value(redisSpanKey).(opentracing.Span)
	if !ok {
		return nil
	}
	defer span.Finish()

	span.SetTag("cmd", cmd.FullName())
	span.LogFields(open_log.Object("args", cmd.Args()))
	if cmd.Err() != nil && cmd.Err() != redis.Nil {
		span.SetTag("error", true)
		span.LogFields(open_log.Error(cmd.Err()))
	}
	return nil
}

func (redisTraceHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	var span opentracing.Span
	parentSpan := opentracing.SpanFromContext(ctx) // 获取父span
	if parentSpan != nil {
		span = opentracing.StartSpan("redis_multi_cmds", opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan("redis_multi_cmds")
	}

	return context.WithValue(ctx, redisSpanKey, span), nil
}

func (redisTraceHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
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
			sp.LogFields(open_log.Error(cmd.Err()))
		}
		sp.SetTag("cmd", cmd.FullName())
		sp.LogFields(open_log.Object("args", cmd.Args()))
		sp.Finish()
	}

	if err != nil {
		span.SetTag("error", true)
		span.LogFields(open_log.Error(err))
	}

	span.LogFields(open_log.Int("cmd_count", len(cmds)))
	return nil
}
