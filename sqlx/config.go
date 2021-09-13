package sqlx

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "sqlx"

// 配置
type SqlxConfig struct {
	Driver           string // 驱动
	Source           string // 连接源
	MaxIdleConns     int    // 最大空闲连接数
	MaxOpenConns     int    // 最大连接池个数
	ConnMaxLifetime  int    // 最大续航时间(毫秒, 0表示无限
}
