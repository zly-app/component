package pulsar_producer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsar/log"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
)

// pulsar 生产者建造者
type IPulsarProducerCreator interface {
	GetPulsarProducer(name ...string) IPulsarProducer
	Close()
}

// Pulsar生产者
type IPulsarProducer interface {
	// topic
	Topic() string
	// 生产者名
	Name() string
	// 同步发送消息
	Send(context.Context, *ProducerMessage) (MessageID, error)
	// 异步发送消息
	SendAsync(context.Context, *ProducerMessage, func(MessageID, *ProducerMessage, error))
	// 刷新客户端中缓存的所有消息，并等待，直到所有消息都成功持久化
	Flush() error
}

type ProducerCreator struct {
	app  core.IApp
	conn *conn.Conn
}

func (p *ProducerCreator) GetPulsarProducer(name ...string) IPulsarProducer {
	return p.conn.GetInstance(p.makeProducer, name...).(IPulsarProducer)
}

func (p *ProducerCreator) Close() { p.conn.CloseAll() }

func (p *ProducerCreator) makeProducer(name string) (conn.IInstance, error) {
	conf := NewConfig()
	err := p.app.GetConfig().ParseComponentConfig(DefaultComponentType, name, conf, true)
	if err == nil {
		err = conf.Check()
	}
	if err != nil {
		return nil, fmt.Errorf("获取组件<%s.%s>配置失败: %v", DefaultComponentType, name, err)
	}

	producer, err := NewProducer(conf)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

// 创建生产者建造者
func NewProducerCreator(app core.IApp) IPulsarProducerCreator {
	p := &ProducerCreator{
		app:  app,
		conn: conn.NewConn(),
	}
	return p
}

var _ IPulsarProducer = (*PulsarProducer)(nil)
var _ conn.IInstance = (*PulsarProducer)(nil)

type PulsarProducer struct {
	client pulsar.Client
	pulsar.Producer
}

func (p *PulsarProducer) Close() {
	p.Producer.Close()
	p.client.Close()
}

// 创建生产者
func NewProducer(conf *Config) (*PulsarProducer, error) {
	co := pulsar.ClientOptions{
		URL:                     conf.Url,
		ConnectionTimeout:       time.Duration(conf.ConnectionTimeout) * time.Millisecond,
		OperationTimeout:        time.Duration(conf.OperationTimeout) * time.Millisecond,
		ListenerName:            conf.ListenerName,
		MaxConnectionsPerBroker: 1,
		Logger:                  log.DefaultNopLogger(),
	}

	client, err := pulsar.NewClient(co)
	if err != nil {
		return nil, fmt.Errorf("创建pulsar客户端失败: %v", err)
	}

	po := pulsar.ProducerOptions{
		Topic: conf.Topic,
		Name:  conf.Name,
		// Properties:                      nil,
		SendTimeout:             -1,
		DisableBlockIfQueueFull: conf.DisableBlockIfQueueFull,
		MaxPendingMessages:      conf.MaxPendingMessages,
		// HashingScheme:                   0,
		CompressionType:  pulsar.NoCompression,
		CompressionLevel: pulsar.Default,
		// MessageRouter:                   nil,
		DisableBatching:         conf.DisableBatching,
		BatchingMaxPublishDelay: time.Duration(conf.BatchingMaxPublishDelay) * time.Millisecond,
		BatchingMaxMessages:     uint(conf.BatchingMaxMessages),
		BatchingMaxSize:         uint(conf.BatchingMaxSize),
		// Interceptors:                    nil,
		// Schema:                          nil,
		PartitionsAutoDiscoveryInterval: time.Duration(conf.PartitionsAutoDiscoveryInterval) * time.Millisecond,
		// Encryption:                      nil,
	}
	if conf.SendTimeout != -1 {
		po.SendTimeout = time.Duration(conf.SendTimeout) * time.Millisecond
	}
	switch strings.ToLower(conf.CompressionType) {
	case "lz4":
		po.CompressionType = pulsar.LZ4
	case "zlib":
		po.CompressionType = pulsar.ZLib
	case "zstd":
		po.CompressionType = pulsar.ZSTD
	}
	switch strings.ToLower(conf.CompressionLevel) {
	case "faster":
		po.CompressionLevel = pulsar.Faster
	case "better":
		po.CompressionLevel = pulsar.Better
	}
	if conf.MaxReconnectToBroker > -1 {
		v := uint(conf.MaxReconnectToBroker)
		po.MaxReconnectToBroker = &v
	}
	if conf.KeyBatcherBuilder {
		po.BatcherBuilderType = pulsar.KeyBasedBatchBuilder
	}

	producer, err := client.CreateProducer(po)
	if err != nil {
		return nil, fmt.Errorf("创建pulsar生产者失败: %v", err)
	}

	return &PulsarProducer{
		client:   client,
		Producer: producer,
	}, nil
}
