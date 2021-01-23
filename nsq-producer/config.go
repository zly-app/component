/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package nsq_producer

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "nsq-producer"

type NsqProducerConfig struct {
	Address      string // 地址: localhost:4150
	AuthSecret   string // 验证秘钥
	ReadTimeout  int64  // 超时(毫秒
	WriteTimeout int64  // 超时(毫秒
	DialTimeout  int64  // 超时(毫秒
}
