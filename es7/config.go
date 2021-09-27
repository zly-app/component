/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package es7

import (
	"errors"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "es7"

const (
	// 默认连接超时(毫秒
	defaultDialTimeout = 5000
	// 默认开启嗅探器
	defaultSniff = false
	// 默认心跳检查
	defaultHealthCheck = true
	// 默认重试次数
	defaultRetry = 0
	// 默认重试间隔(毫秒
	defaultRetryInterval = 500
	//  默认启用gzip压缩
	defaultGZip = false
)

// es7配置
type ES7Config struct {
	Address       string // 地址: http://localhost1:9200,http://localhost2:9200
	UserName      string // 用户名
	Password      string // 密码
	DialTimeout   int64  // 连接超时(毫秒
	Sniff         bool   // 开启嗅探器
	HealthCheck   bool   // 心跳检查(默认true
	Retry         int    // 重试次数
	RetryInterval int    // 重试间隔(毫秒)
	GZip          bool   // 启用gzip压缩
}

func newConfig() *ES7Config {
	return &ES7Config{
		Sniff:         defaultSniff,
		HealthCheck:   defaultHealthCheck,
		Retry:         defaultRetry,
		RetryInterval: defaultRetryInterval,
		GZip:          defaultGZip,
	}
}

func (conf *ES7Config) Check() error {
	if conf.Address == "" {
		return errors.New("es7的address为空")
	}
	if conf.DialTimeout < 1 {
		conf.DialTimeout = defaultDialTimeout
	}
	return nil
}
