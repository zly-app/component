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
	defaultMaxIdleConns = 3
	// 默认最大连接池个数
	defaultMaxOpenConns = 10
	// 默认最大续航时间
	defaultConnMaxLifetime = 0
	// 默认表映射规则
	defaultTableMapperRule = "GonicMapper"
	// 默认列映射规则
	defaultColumnMapperRule = "GonicMapper"
)

// xorm配置
type XormConfig struct {
	Driver           string // 驱动
	Source           string // 连接源
	MaxIdleConns     int    // 最大空闲连接数
	MaxOpenConns     int    // 最大连接池个数
	ConnMaxLifetime  int    // 最大续航时间(毫秒, 0表示无限
	TableMapperRule  string // 表映射规则, 可选 SnakeMapper, SameMapper, GonicMapper, 默认为 GonicMapper
	ColumnMapperRule string // 列映射规则, 可选 SnakeMapper, SameMapper, GonicMapper, 默认为 GonicMapper
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
	if conf.ConnMaxLifetime < 1 {
		conf.ConnMaxLifetime = defaultConnMaxLifetime
	}
	if conf.TableMapperRule == "" {
		conf.TableMapperRule = defaultTableMapperRule
	}
	if conf.ColumnMapperRule == "" {
		conf.ColumnMapperRule = defaultColumnMapperRule
	}
	return nil
}
