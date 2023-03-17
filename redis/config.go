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
)

const (
	defMinIdle         = 2
	defMaxIdle         = defMinIdle * 2
	defPoolSize        = 10
	defIdleTimeout     = 3600
	defWaitTimeout     = 5
	defConnectTimeout  = 5
	defMaxConnLifetime = 3600
	defReadTimeoutSec  = 5
	defWriteTimeoutSec = 5
)

// redis配置
type RedisConfig struct {
	Address  string // 地址: host1:port1,host2:port2
	UserName string // 用户名
	Password string // 密码
	DB       int    // db, 只有非集群有效

	MinIdle         int // 最小闲置连接数
	MaxIdle         int // 最大闲置连接数
	PoolSize        int // 客户端池大小
	IdleTimeout     int // 空闲链接超时时间, 单位秒, 如果一个连接长时间未使用将被视为连接无效, 小于1表示永不超时
	WaitTimeout     int // 等待获取连接的超时时间, 单位秒
	ConnectTimeout  int // 连接超时, 单位秒
	MaxConnLifetime int // 一个连接最大存活时间, 单位秒, 小于1表示不限制

	MaxRetries      int // 操作尝试次数, <1 表示不重试
	ReadTimeoutSec  int // 超时, 秒
	WriteTimeoutSec int // 超时, 秒
}

func newRedisConfig() *RedisConfig {
	return &RedisConfig{
		ReadTimeoutSec:  defReadTimeoutSec,
		WriteTimeoutSec: defWriteTimeoutSec,
		IdleTimeout:     defIdleTimeout,
		MaxConnLifetime: defMaxConnLifetime,
	}
}

func (conf *RedisConfig) Check() error {
	if conf.Address == "" {
		return errors.New("redis的address为空")
	}

	if conf.MinIdle < 1 {
		conf.MinIdle = defMinIdle
	}
	if conf.MaxIdle < 1 {
		conf.MaxIdle = defMaxIdle
	}
	if conf.MaxIdle < conf.MinIdle {
		conf.MaxIdle = conf.MinIdle * 2
	}
	if conf.PoolSize < 1 {
		conf.PoolSize = defPoolSize
	}
	if conf.IdleTimeout < 1 {
		conf.IdleTimeout = 0
	}
	if conf.WaitTimeout < 1 {
		conf.WaitTimeout = defWaitTimeout
	}
	if conf.ConnectTimeout < 1 {
		conf.ConnectTimeout = defConnectTimeout
	}
	if conf.MaxConnLifetime < 1 {
		conf.MaxConnLifetime = 0
	}

	if conf.MaxRetries < 1 {
		conf.MaxRetries = -1 // -1 才是禁用
	}
	return nil
}
