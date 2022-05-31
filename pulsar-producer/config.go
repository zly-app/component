package pulsar_producer

import (
	"fmt"
	"strings"

	"github.com/zly-app/zapp/core"
)

const DefaultComponentType core.ComponentType = "pulsar-producer"

const (
	// pulsar地址
	defUrl = "pulsar://localhost:6650"
	// 连接超时
	defConnectionTimeout = 5000
	// 操作超时
	defOperationTimeout = 30000

	// topic
	defTopic = "persistent://public/default/test"
	// 发送超时
	defSendTimeout = 30000
	// 是否关闭队列阻塞
	defDisableBlockIfQueueFull = false
	// 待确认消息队列大小
	defMaxPendingMessages = 1000
	// 压缩器类型
	defCompressionType = "lz4"
	// 压缩等级
	defCompressionLevel = "default"
	// 是否关闭批处理
	defDisableBatching = false
	// 是否启用基于key的批次构建容器
	defKeyBatcherBuilder = false
	// 批处理等待时间
	defBatchingMaxPublishDelay = 10
	// 单批次最大消息数量
	defBatchingMaxMessages = 1000
	// 单批次消息最大占用字节大小
	defBatchingMaxSize = 128 << 10
	// 重新连接到broker的最大次数
	defMaxReconnectToBroker = -1
	// 发现新分区的时间间隔
	defPartitionsAutoDiscoveryInterval = 60000
)

type Config struct {
	Url               string // pulsar地址, 示例: pulsar://localhost:6600,localhost:6650
	ListenerName      string // pulsar使用的监听器名, 示例: external
	ConnectionTimeout int    // 连接超时, 单位毫秒
	OperationTimeout  int    // 操作超时, 单位毫秒

	Topic                           string // topic, 示例: persistent://public/default/test
	Name                            string // 生产者名, 如果为空则自动生成一个全局唯一名称
	SendTimeout                     int    // 发送超时, 单位毫秒. -1表示禁用
	DisableBlockIfQueueFull         bool   // 是否关闭队列阻塞, 如果设为false, 在队列满时发送消息会阻塞, 如果设为true, 在队列满时发送消息会返回err, 默认 false
	MaxPendingMessages              int    // 保存等待从代理接收确认消息的消息的队列的最大大小
	CompressionType                 string // 压缩器类型, 支持 none, lz4, zlib, zstd
	CompressionLevel                string // 压缩等级, 支持 default, faster, better
	DisableBatching                 bool   // 是否关闭批处理, 如果消费者使用KeyShared模式必须关闭或者启用基于key的批次构建容器, 否则可能导致相同的key被不同的消费者消费
	KeyBatcherBuilder               bool   // 是否启用基于key的批次构建容器
	BatchingMaxPublishDelay         int    // 批处理等待时间, 单位毫秒
	BatchingMaxMessages             int    // 单批次消息最大数量
	BatchingMaxSize                 int    // 单批次消息最大占用字节大小
	MaxReconnectToBroker            int    // 重新连接到broker的最大次数, -1表示不限
	PartitionsAutoDiscoveryInterval int    // 发现新分区的时间间隔, 单位毫秒
}

func NewConfig() *Config {
	return &Config{
		DisableBlockIfQueueFull: defDisableBlockIfQueueFull,
		DisableBatching:         defDisableBatching,
		KeyBatcherBuilder:       defKeyBatcherBuilder,
		MaxReconnectToBroker:    defMaxReconnectToBroker,
	}
}

func (conf *Config) Check() error {
	if conf.Url == "" {
		conf.Url = defUrl
	}
	if conf.ConnectionTimeout < 1 {
		conf.ConnectionTimeout = defConnectionTimeout
	}
	if conf.OperationTimeout < 1 {
		conf.OperationTimeout = defOperationTimeout
	}

	if conf.Topic == "" {
		conf.Topic = defTopic
	}
	if conf.SendTimeout < 0 {
		conf.SendTimeout = -1
	} else if conf.SendTimeout == 0 {
		conf.SendTimeout = defSendTimeout
	}
	if conf.MaxPendingMessages < 1 {
		conf.MaxPendingMessages = defMaxPendingMessages
	}
	switch strings.ToLower(conf.CompressionType) {
	case "":
		conf.CompressionType = defCompressionType
	case "none", "lz4", "zlib", "zstd":
	default:
		return fmt.Errorf("不支持的压缩器类型: %v", conf.CompressionType)
	}
	switch strings.ToLower(conf.CompressionLevel) {
	case "":
		conf.CompressionLevel = defCompressionLevel
	case "default", "faster", "better":
	default:
		return fmt.Errorf("不支持的压缩等级: %v", conf.CompressionLevel)
	}
	if conf.BatchingMaxPublishDelay < 1 {
		conf.BatchingMaxPublishDelay = defBatchingMaxPublishDelay
	}
	if conf.BatchingMaxMessages < 1 {
		conf.BatchingMaxMessages = defBatchingMaxMessages
	}
	if conf.BatchingMaxSize < 1 {
		conf.BatchingMaxSize = defBatchingMaxSize
	}
	if conf.MaxReconnectToBroker < 0 {
		conf.MaxReconnectToBroker = defMaxReconnectToBroker
	}
	if conf.PartitionsAutoDiscoveryInterval < 1 {
		conf.PartitionsAutoDiscoveryInterval = defPartitionsAutoDiscoveryInterval
	}
	return nil
}
