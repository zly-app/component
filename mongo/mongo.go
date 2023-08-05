package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 创建一个客户端
func NewClient(conf *MongoConfig) (*Client, error) {
	err := conf.Check()
	if err != nil {
		return nil, fmt.Errorf("mongo客户端配置错误: %v", err)
	}

	connectTimeout := time.Duration(conf.ConnectTimeout) * time.Second
	timeout := time.Duration(conf.DoTimeoutSec) * time.Second
	idleTimeout := time.Duration(conf.IdleTimeout) * time.Second
	opt := &options.ClientOptions{
		Hosts:           strings.Split(conf.Address, ","),
		MinPoolSize:     &conf.MinPoolSize,
		MaxPoolSize:     &conf.MaxPoolSize,
		ConnectTimeout:  &connectTimeout,
		Timeout:         &timeout,
		MaxConnIdleTime: &idleTimeout,
		Direct:          &conf.Direct,
	}
	if conf.UserName != "" {
		opt.Auth = &options.Credential{
			AuthSource:  conf.DB,
			Username:    conf.UserName,
			Password:    conf.Password,
			PasswordSet: true,
		}
	}

	client, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		return nil, err
	}
	return client, err
}
