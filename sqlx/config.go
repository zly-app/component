package sqlx

import (
	"errors"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "sqlx"

const (
	// 默认最大空闲连接数
	defaultMaxIdleConns = 3
	// 默认最大连接池个数
	defaultMaxOpenConns = 10
	// 默认最大续航时间
	defaultConnMaxLifetime = 0
)

// 配置
type SqlxConfig struct {
	Driver          string // 驱动
	Source          string // 连接源
	MaxIdleConns    int    // 最大空闲连接数
	MaxOpenConns    int    // 最大连接池个数
	ConnMaxLifetime int    // 最大续航时间(毫秒, 0表示无限
}

func newConfig() *SqlxConfig {
	return &SqlxConfig{}
}

func (conf *SqlxConfig) Check() error {
	if conf.Driver == "" {
		return errors.New("xorm的Driver为空")
	}
	if conf.Source == "" {
		return errors.New("xorm的Source为空")
	}
	if conf.MaxIdleConns < 1 {
		conf.MaxOpenConns = defaultMaxIdleConns
	}
	if conf.MaxOpenConns < 1 {
		conf.MaxOpenConns = defaultMaxOpenConns
	}
	if conf.ConnMaxLifetime < 1 {
		conf.ConnMaxLifetime = defaultConnMaxLifetime
	}
	return nil
}
