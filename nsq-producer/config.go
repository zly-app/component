/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package nsq_producer

import (
	"errors"

	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "nsq-producer"

const (
	// 默认读取超时
	defaultReadTimeout = 30000
	// 默认写入超时
	defaultWriteTimeout = 5000
	// 默认连接超时
	defaultDialTimeout = 2000
)

type NsqProducerConfig struct {
	Address           string // 地址: localhost:4150
	AuthSecret        string // 验证秘钥
	ReadTimeout       int64  // 超时(毫秒
	WriteTimeout      int64  // 超时(毫秒
	DialTimeout       int64  // 超时(毫秒
}

func newConfig() *NsqProducerConfig {
	return &NsqProducerConfig{
		Address:           "",
		AuthSecret:        "",
		ReadTimeout:       defaultReadTimeout,
		WriteTimeout:      defaultWriteTimeout,
		DialTimeout:       defaultDialTimeout,
	}
}

func (conf *NsqProducerConfig) Check() error {
	if conf.ReadTimeout <= 0 {
		conf.ReadTimeout = defaultReadTimeout
	}
	if conf.WriteTimeout <= 0 {
		conf.WriteTimeout = defaultWriteTimeout
	}
	if conf.DialTimeout <= 0 {
		conf.DialTimeout = defaultDialTimeout
	}

	if conf.Address == "" {
		return errors.New("address为空")
	}

	return nil
}
