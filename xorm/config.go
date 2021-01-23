/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/21
   Description :
-------------------------------------------------
*/

package xorm

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "xorm"

// xorm配置
type XormConfig struct {
	Driver          string // 驱动
	Source          string // 连接源
	MaxIdleConns    int    // 最大空闲连接数
	MaxOpenConns    int    // 最大连接池个数
	ConnMaxLifetime int    // 最大续航时间(毫秒, 0表示无限
}
