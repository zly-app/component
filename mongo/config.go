package mongo

import (
	"errors"
)

const (
	defMinPoolSize    = 2
	defMaxPoolSize    = 5
	defIdleTimeout    = 3600
	defConnectTimeout = 5
	defDoTimeoutSec   = 5
)

type MongoConfig struct {
	Address  string // 地址: host1:port1,host2:port2
	UserName string // 用户名
	Password string // 密码
	DB       string // db

	MinPoolSize uint64 // 最小连接池
	MaxPoolSize uint64 // 最大连接池

	IdleTimeout    int // 空闲链接超时时间, 单位秒, 如果一个连接长时间未使用将被视为连接无效, 小于1表示永不超时
	ConnectTimeout int // 连接超时, 单位秒
	DoTimeoutSec   int // 一个操作的超时, 秒
}

func NewMongoConfig() *MongoConfig {
	return &MongoConfig{
		MinPoolSize:    defMinPoolSize,
		MaxPoolSize:    defMaxPoolSize,
		IdleTimeout:    defIdleTimeout,
		ConnectTimeout: defConnectTimeout,
		DoTimeoutSec:   defDoTimeoutSec,
	}
}

func (conf *MongoConfig) Check() error {
	if conf.Address == "" {
		return errors.New("mongo的address为空")
	}

	if conf.MinPoolSize < 1 {
		conf.MinPoolSize = defMinPoolSize
	}
	if conf.MaxPoolSize < 1 {
		conf.MaxPoolSize = defMaxPoolSize
	}
	if conf.IdleTimeout < 1 {
		conf.IdleTimeout = defIdleTimeout
	}
	if conf.ConnectTimeout < 1 {
		conf.ConnectTimeout = defConnectTimeout
	}
	if conf.DoTimeoutSec < 1 {
		conf.DoTimeoutSec = defDoTimeoutSec
	}
	return nil
}
