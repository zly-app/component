/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package redis

import (
	"errors"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "redis"

const (
	// 默认最小空闲连接数
	defaultMinIdleConns = 1
	// 默认客户端池大小
	defaultPoolSize = 1
	// 默认读取超时
	defaultReadTimeout = 5000
	// 默认写入超时
	defaultWriteTimeout = 5000
	// 默认连接超时
	defaultDialTimeout = 5000
	// 默认启用开放链路追踪
	defaultEnableOpenTrace = false
)

// redis配置
type RedisConfig struct {
	Address         string // 地址: host1:port1,host2:port2
	UserName        string // 用户名
	Password        string // 密码
	DB              int    // db, 只有非集群有效
	IsCluster       bool   // 是否为集群
	MinIdleConns    int    // 最小空闲连接数
	PoolSize        int    // 客户端池大小
	ReadTimeout     int64  // 超时(毫秒
	WriteTimeout    int64  // 超时(毫秒
	DialTimeout     int64  // 超时(毫秒
	EnableOpenTrace bool   // 启用开放链路追踪
}

func newRedisConfig() *RedisConfig {
	return &RedisConfig{
		MinIdleConns:    defaultMinIdleConns,
		PoolSize:        defaultPoolSize,
		ReadTimeout:     defaultReadTimeout,
		WriteTimeout:    defaultWriteTimeout,
		DialTimeout:     defaultDialTimeout,
		EnableOpenTrace: defaultEnableOpenTrace,
	}
}

func (conf *RedisConfig) Check() error {
	if conf.Address == "" {
		return errors.New("redis的address为空")
	}
	return nil
}
