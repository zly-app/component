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
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "grpc-client"

// grpc客户端配置
type GrpcClientConfig struct {
	Address     string // 链接地址
	Registry    string // 注册器, 默认为 local
	Balance     string // 负载均衡, 默认为 round_robin
	DialTimeout int    // 连接超时(毫秒), 默认为 5000
}
