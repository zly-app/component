/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/1
   Description :
-------------------------------------------------
*/

package ssdb

import (
	"errors"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "ssdb"

const (
	// 默认最小空闲连接数
	defaultMinIdleConns = 3
	// 默认最大连接池个数
	defaultPoolSize = 10
	// 默认获取客户端超时(毫秒)
	defaultGetClientTimeout = 5000
	// 默认获取客户端超时等待队列大小
	defaultGetClientWaitQueueSize = 1000
	// 默认读取超时
	defaultReadTimeout = 5000
	// 默认写入超时
	defaultWriteTimeout = 5000
	// 默认连接超时
	defaultDialTimeout = 5000
	// 默认是否启用重试
	defaultRetryEnabled = false
)

type SsdbConfig struct {
	Address                string // 地址: host1:port1
	Password               string // 密码
	MinIdleConns           int    // 最小连接池数
	PoolSize               int    // 最大连接池个数
	GetClientTimeout       int    // 获取客户端超时(毫秒)
	GetClientWaitQueueSize int    // 获取客户端超时等待队列大小, 当连接池满后, 新建连接将等待池中连接释放后才可以继续, 超出队列的会抛出异常
	ReadTimeout            int    // 读取超时(毫秒
	WriteTimeout           int    // 写入超时(毫秒
	DialTimeout            int    // 连接超时(毫秒
	RetryEnabled           bool   // 是否启用重试, 设置为true时, 如果请求失败会再重试一次
}

func newSsdbConfig() *SsdbConfig {
	return &SsdbConfig{
		MinIdleConns:           defaultMinIdleConns,
		PoolSize:               defaultPoolSize,
		GetClientTimeout:       defaultGetClientTimeout,
		GetClientWaitQueueSize: defaultGetClientWaitQueueSize,
		ReadTimeout:            defaultReadTimeout,
		WriteTimeout:           defaultWriteTimeout,
		DialTimeout:            defaultDialTimeout,
		RetryEnabled:           defaultRetryEnabled,
	}
}

func (conf *SsdbConfig) Check() error {
	if conf.Address == "" {
		return errors.New("ssdb的address为空")
	}
	return nil
}
