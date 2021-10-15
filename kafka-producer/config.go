package kafka_producer

import (
	"errors"

	"github.com/Shopify/sarama"
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "kafka-producer"

const (
	// 默认kafka版本
	defaultKafkaVersion = "2.0.0"
	// 默认读取超时
	defaultReadTimeout = 10000
	// 默认写入超时
	defaultWriteTimeout = 10000
	// 默认连接超时
	defaultDialTimeout = 2000
	// 默认消息的最大允许大小
	defaultMaxMessageBytes = 1048576
	// 默认压缩类型
	defaultCompression = "none"
	// 默认启用Successes通道
	defaultEnabledSuccessesChannel = true
	// 默认启用Errors通道
	defaultEnabledErrorsChannel = true
	// 默认发送失败重试次数
	defaultRetryCount = 3
	// 默认发送失败重试间隔时间
	defaultRetryInterval = 100
	// 默认通道缓冲数
	defaultChannelBufferSize = 256
)

type Config struct {
	Address                 string // 地址, 多个地址用半角逗号连接
	KafkaVersion            string // kafka版本, 示例: v2.0.0, 2.0.0
	ReadTimeout             int64  // 超时(毫秒)
	WriteTimeout            int64  // 超时(毫秒)
	DialTimeout             int64  // 超时(毫秒)
	MaxMessageBytes         int    // 消息的最大允许大小(字节)
	Compression             string // 压缩类型, none, gzip, snappy, lz4, zstd
	CompressionLevel        int    // 压缩级别, 根据编码不同具有不同的区间, 默认为不同编码默认级别
	EnabledSuccessesChannel bool   // 启用Successes通道, 如果启用, 必须循环从这个通道读取数据以防止死锁. 该选项仅异步生产者有效(默认启用)
	EnabledErrorsChannel    bool   // 启用Errors通道, 如果启用, 必须循环从这个通道读取数据以防止死锁. 该选项仅异步生产者有效(默认启用)
	RetryCount              int    // 发送失败重试次数
	RetryInterval           int    // 发送失败重试间隔时间(毫秒)
	ChannelBufferSize       int    // 通道缓冲数, 要在内部和外部通道中缓冲的事件数量
	ProxyAddress            string // 代理地址, 支持 socks5, socks5h
	ProxyUser               string // 代理用户名
	ProxyPassword           string // 代理用户的密码

	kConf *sarama.Config
}

func newConfig() *Config {
	return &Config{
		KafkaVersion:            defaultKafkaVersion,
		CompressionLevel:        sarama.CompressionLevelDefault,
		EnabledSuccessesChannel: defaultEnabledSuccessesChannel,
		EnabledErrorsChannel:    defaultEnabledErrorsChannel,
		RetryCount:              defaultRetryCount,
	}
}

func (conf *Config) Check() error {
	if conf.Address == "" {
		return errors.New("address为空")
	}
	if conf.ReadTimeout <= 0 {
		conf.ReadTimeout = defaultReadTimeout
	}
	if conf.WriteTimeout <= 0 {
		conf.WriteTimeout = defaultWriteTimeout
	}
	if conf.DialTimeout <= 0 {
		conf.DialTimeout = defaultDialTimeout
	}
	if conf.MaxMessageBytes <= 0 {
		conf.MaxMessageBytes = defaultMaxMessageBytes
	}
	if conf.Compression == "" {
		conf.Compression = defaultCompression
	}
	if conf.RetryInterval <= 0 {
		conf.RetryInterval = defaultRetryInterval
	}
	if conf.ChannelBufferSize <= 0 {
		conf.ChannelBufferSize = defaultChannelBufferSize
	}
	return nil
}
