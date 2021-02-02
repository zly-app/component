/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/2/1
   Description :
-------------------------------------------------
*/

package ssdb

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "ssdb"

type SsdbConfig struct {
	Host                   string // host
	Port                   int    // port
	Password               string // 密码
	GetClientTimeout       int    // 获取客户端超时(毫秒)
	GetClientWaitQueueSize int    // 获取客户端超时等待队列大小, 当连接池满后, 新建连接将等待池中连接释放后才可以继续, 超出队列的会抛出异常
	ReadTimeout            int    // 读取超时(毫秒
	WriteTimeout           int    // 写入超时(毫秒
	DialTimeout            int    // 连接超时(毫秒
	MinPoolSize            int    // 最小连接池数
	MaxPoolSize            int    // 最大连接池个数
	RetryEnabled           bool   // 是否启用重试, 设置为true时, 如果请求失败会再重试一次
}
