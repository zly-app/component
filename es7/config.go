/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package es7

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "es7"

// es7配置
type ES7Config struct {
	Address       string // 地址: http://localhost1:9200,http://localhost2:9200
	UserName      string // 用户名
	Password      string // 密码
	DialTimeout   int64  // 连接超时(毫秒
	Sniff         bool   // 开启嗅探器
	Healthcheck   *bool  // 心跳检查(默认true
	Retry         int    // 重试次数
	RetryInterval int    // 重试间隔(毫秒)
	GZip          bool   // 启用gzip压缩
}
