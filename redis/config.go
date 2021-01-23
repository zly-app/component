/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package redis

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "redis"

// redis配置
type RedisConfig struct {
	Address      string // 地址: host1:port1,host2:port2
	UserName     string // 用户名
	Password     string // 密码
	DB           int    // db, 只有非集群有效
	IsCluster    bool   // 是否为集群
	MinIdleConns int    // 最小空闲连接数
	PoolSize     int    // 客户端池大小
	ReadTimeout  int64  // 超时(毫秒
	WriteTimeout int64  // 超时(毫秒
	DialTimeout  int64  // 超时(毫秒
}
