/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/23
   Description :
-------------------------------------------------
*/

package grpc_client

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/opentracing/opentracing-go"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/zly-app/component/grpc-client/balance/round_robin"
	"github.com/zly-app/component/grpc-client/registry/local"
)

var typeOfGrpcClientConn = reflect.TypeOf((*grpc.ClientConn)(nil))
var typeOfGrpcClientConnInterface = reflect.TypeOf((*grpc.ClientConnInterface)(nil)).Elem()

type IGrpcClient interface {
	// 注册grpc客户端创造者, 这个方法应该在app.Run之前调用
	//
	// creator 必须实现 func (cc *grpc.ClientConn) Any 或 func(cc grpc.ClientConnInterface) Any
	// 示例:
	//     c.RegistryGrpcClientCreator("xxx", pb.NewXXXClient)
	RegistryGrpcClientCreator(name string, creator interface{})
	// 获取grpc客户端, 如果未注册grpc客户端建造者会panic
	GetGrpcClient(name string) interface{}
	// 关闭客户端
	Close()
}

type instance struct {
	cc     *grpc.ClientConn
	client interface{}
}

func (i *instance) Close() {
	_ = i.cc.Close()
}

type GrpcClient struct {
	app           core.IApp
	conn          *conn.Conn
	componentType core.ComponentType

	configs    map[string]*GrpcClientConfig
	creatorMap map[string]reflect.Value
}

func NewGrpcClient(app core.IApp, componentType ...core.ComponentType) IGrpcClient {
	g := &GrpcClient{
		app:           app,
		conn:          conn.NewConn(),
		componentType: DefaultComponentType,

		configs:    make(map[string]*GrpcClientConfig),
		creatorMap: make(map[string]reflect.Value),
	}
	if len(componentType) > 0 {
		g.componentType = componentType[0]
	}

	// 加载配置
	err := app.GetConfig().Parse("components."+string(g.componentType), &g.configs)
	if err != nil {
		app.Fatal("解析组件配置失败", zap.String("componentType", string(g.componentType)), zap.Error(err))
	}
	// 检查配置
	for _, conf := range g.configs {
		if err := conf.Check(); err != nil {
			app.Fatal("组件配置错误", zap.Error(err))
		}
	}

	// 分析配置
	for name, conf := range g.configs {
		switch conf.Registry {
		case local.Name:
			local.RegistryAddress(name, conf.Address)
		default:
			logger.Log.Fatal("未定义的Grpc注册器", zap.String("registry", conf.Registry))
		}

		_ = g.getBalance(conf.Balance)
	}

	return g
}

func (g *GrpcClient) RegistryGrpcClientCreator(name string, creator interface{}) {
	createType := reflect.TypeOf(creator)
	if createType.Kind() != reflect.Func {
		logger.Log.Fatal("grpc客户端建造者必须是函数")
		return
	}

	if createType.NumIn() != 1 {
		logger.Log.Fatal("grpc客户端建造者入参为1个")
		return
	}

	arg0 := createType.In(0)
	if !arg0.AssignableTo(typeOfGrpcClientConn) && !arg0.AssignableTo(typeOfGrpcClientConnInterface) {
		logger.Log.Fatal("grpc客户端建造者入参类型必须是 *grpc.ClientConn 或 grpc.ClientConnInterface")
		return
	}

	if createType.NumOut() != 1 {
		logger.Log.Fatal("grpc客户端建造者必须有一个返回值")
		return
	}

	g.creatorMap[name] = reflect.ValueOf(creator)
}

func (g *GrpcClient) GetGrpcClient(name string) interface{} {
	return g.conn.GetInstance(g.makeClient, name).(*instance).client
}

func (g *GrpcClient) makeClient(name string) (conn.IInstance, error) {
	// 获取配置
	conf, ok := g.configs[name]
	if !ok {
		return nil, errors.New("试图获取未注册的grpc客户端")
	}

	// 获取建造者
	creator, ok := g.creatorMap[name]
	if !ok {
		return nil, errors.New("未注册grpc客户端建造者")
	}

	cc, err := g.makeConn(name, conf)
	if err != nil {
		return nil, fmt.Errorf("构建conn错误: name=%s, registry=%s, balance=%s, err: %s", name, conf.Registry, conf.Balance, err)
	}

	client := creator.Call([]reflect.Value{reflect.ValueOf(cc)})[0].Interface()

	return &instance{
		cc:     cc,
		client: client,
	}, nil
}

func (g *GrpcClient) makeConn(name string, conf *GrpcClientConfig) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(g.app.BaseContext(), time.Duration(conf.DialTimeout)*time.Millisecond)
	defer cancel()

	target := conf.Registry + ":///" + name
	opts := []grpc.DialOption{
		g.getBalance(conf.Balance), // 均衡器
		grpc.WithBlock(),           // 等待连接成功. 注意, 这个不要作为配置项, 因为要和zapp的conn组件配合, 所以它是必须的.
	}

	var chainUnaryClientList []grpc.UnaryClientInterceptor

	if *conf.InsecureDial {
		opts = append(opts, grpc.WithInsecure()) // 不安全连接
	}
	if *conf.EnableOpenTrace {
		chainUnaryClientList = append(chainUnaryClientList, UnaryClientOpenTraceInterceptor)
	}
	opts = append(opts, grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(chainUnaryClientList...)))
	return grpc.DialContext(ctx, target, opts...)
}

func (g *GrpcClient) getBalance(balance string) grpc.DialOption {
	switch balance {
	case round_robin.Name:
		return round_robin.Balance()
	default:
		logger.Log.Fatal("未定义的Grpc客户端均衡器", zap.String("balance", balance))
	}
	return nil
}

func (g *GrpcClient) Close() {
	g.conn.CloseAll()
}

type TextMapCarrier struct {
	metadata.MD
}

func (t TextMapCarrier) Set(key, val string) {
	t.MD[key] = append(t.MD[key], val)
}

// 开放链路追踪hook
func UnaryClientOpenTraceInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	var span opentracing.Span
	parentSpan := opentracing.SpanFromContext(ctx) // 获取父span
	if parentSpan != nil {
		span = opentracing.StartSpan(method, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = opentracing.StartSpan(method)
	}
	defer span.Finish()

	// 取出元数据
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// 如果对元数据修改必须使用它的副本
		md = md.Copy()
	} else {
		md = metadata.New(nil)
	}

	// 注入
	carrier := TextMapCarrier{md}
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		logger.Log.Error("grpc trace inject err", zap.Error(err))
	} else {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	span.SetTag("target", cc.Target())
	span.SetTag("req", req)
	err = invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("err", err.Error())
	} else {
		span.SetTag("reply", reply)
	}
	return err
}
