/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/23
   Description :
-------------------------------------------------
*/

package grpc_client

import (
	"github.com/zly-app/zapp/core"

	"github.com/zly-app/component/grpc-client/balance/round_robin"
	"github.com/zly-app/component/grpc-client/registry/static"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "grpc-client"

const (
	// 默认地址
	defaultAddress = "127.0.0.1:3000"
	// 默认注册器
	defaultRegistry = static.Name
	// 默认均衡器
	defaultBalance = round_robin.Name
	// 默认连接超时
	defaultDialTimeout = 5000
	// 默认不安全的连接
	defaultInsecureDial = true
	// 默认启用开放链路追踪
	defaultEnableOpenTrace = true
)

// grpc客户端配置
type GrpcClientConfig struct {
	Address           string // 链接地址
	Registry          string // 注册器, 默认为 static
	Balance           string // 负载均衡, 默认为 round_robin
	DialTimeout       int    // 连接超时(毫秒), 默认为 5000
	InsecureDial      bool   // 不安全的连接
	EnableOpenTrace   bool   // 启用开放链路追踪
	ReqLogLevelIsInfo bool   // 请求日志等级设为info
	RspLogLevelIsInfo bool   // 响应日志等级设为info
}

func newConfig() *GrpcClientConfig {
	return &GrpcClientConfig{
		InsecureDial:    defaultInsecureDial,
		EnableOpenTrace: defaultEnableOpenTrace,
	}
}

func (conf *GrpcClientConfig) Check() {
	if conf.Address == "" {
		conf.Address = defaultAddress
	}
	if conf.Registry == "" {
		conf.Registry = defaultRegistry
	}
	if conf.Balance == "" {
		conf.Balance = defaultBalance
	}
	if conf.DialTimeout < 1 {
		conf.DialTimeout = defaultDialTimeout
	}
}
