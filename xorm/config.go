/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/21
   Description :
-------------------------------------------------
*/

package xorm

import (
	"errors"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "xorm"

const (
	// 默认最大空闲连接数
	defaultMaxIdleConns = 2
	// 默认最大连接池个数
	defaultMaxOpenConns = 5
	// 默认最大续航时间
	defaultConnMaxLifetimeSec = 0
	// 默认表映射规则
	defaultTableMapperRule = "GonicMapper"
	// 默认列映射规则
	defaultColumnMapperRule = "GonicMapper"
	// 默认时区
	defTZ = "Asia/Shanghai"
)

// xorm配置
type XormConfig struct {
	Driver             string // 驱动
	Source             string // 连接源
	MaxIdleConns       int    // 最大空闲连接数
	MaxOpenConns       int    // 最大连接池个数
	ConnMaxLifetimeSec int    // 最大续航时间, 秒, 0表示无限
	TableMapperRule    string // 表映射规则, 可选 SnakeMapper, SameMapper, GonicMapper, 默认为 GonicMapper
	ColumnMapperRule   string // 列映射规则, 可选 SnakeMapper, SameMapper, GonicMapper, 默认为 GonicMapper
	TZ                 string // 时区
}

func newConfig() *XormConfig {
	return &XormConfig{}
}

func (conf *XormConfig) Check() error {
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
	if conf.ConnMaxLifetimeSec < 1 {
		conf.ConnMaxLifetimeSec = defaultConnMaxLifetimeSec
	}
	if conf.TableMapperRule == "" {
		conf.TableMapperRule = defaultTableMapperRule
	}
	if conf.ColumnMapperRule == "" {
		conf.ColumnMapperRule = defaultColumnMapperRule
	}
	if conf.TZ == "" {
		conf.TZ = defTZ
	}
	return nil
}
