package kafka_producer

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/core"
	"golang.org/x/net/proxy"
)

type IKafkaProducer interface {
	// 获取kafka同步生产者
	GetKafkaSyncProducer(name ...string) SyncProducer
	// 获取kafka异步生产者
	GetKafkaAsyncProducer(name ...string) AsyncProducer
	// 关闭
	Close()
}

// 同步生产者
type SyncProducer interface {
	SendMessage(msg *ProducerMessage) (partition int32, offset int64, err error)
	SendMessages(msgs []*ProducerMessage) error
}

// 同步实例
type syncInstance struct {
	syncProducer sarama.SyncProducer
}

func (i *syncInstance) Close() {
	if i.syncProducer != nil {
		_ = i.syncProducer.Close()
	}
}

// 异步生产者
type AsyncProducer interface {
	Input() chan<- *ProducerMessage
	Successes() <-chan *ProducerMessage
	Errors() <-chan *ProducerError
}

// 异步实例
type asyncInstance struct {
	asyncProducer sarama.AsyncProducer
}

func (i *asyncInstance) Close() {
	if i.asyncProducer != nil {
		_ = i.asyncProducer.Close()
	}
}

type KafkaProducer struct {
	app           core.IApp
	connSync      *conn.Conn
	connAsync     *conn.Conn
	componentType core.ComponentType
}

func NewKafkaProducer(app core.IApp, componentType ...core.ComponentType) IKafkaProducer {
	k := &KafkaProducer{
		app:           app,
		connSync:      conn.NewConn(),
		connAsync:     conn.NewConn(),
		componentType: DefaultComponentType,
	}
	if len(componentType) > 0 {
		k.componentType = componentType[0]
	}
	return k
}

func (k *KafkaProducer) GetKafkaSyncProducer(name ...string) SyncProducer {
	return k.connSync.GetInstance(k.makeSyncClient, name...).(*syncInstance).syncProducer
}

func (k *KafkaProducer) GetKafkaAsyncProducer(name ...string) AsyncProducer {
	return k.connAsync.GetInstance(k.makeAsyncClient, name...).(*asyncInstance).asyncProducer
}

// 生成配置
func (k *KafkaProducer) makeConf(name string) (*Config, error) {
	conf := newConfig()
	err := k.app.GetConfig().ParseComponentConfig(k.componentType, name, conf)
	if err != nil {
		return nil, err
	}
	if err = conf.Check(); err != nil {
		return nil, fmt.Errorf("组件%s的配置错误: %s", k.componentType, err)
	}

	kConf := sarama.NewConfig()

	kConf.Net.DialTimeout = time.Duration(conf.DialTimeout) * time.Millisecond
	kConf.Net.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Millisecond
	kConf.Net.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Millisecond
	if conf.ProxyAddress != "" {
		kConf.Net.Proxy.Enable = true
		var auth *proxy.Auth
		if conf.ProxyUser != "" {
			auth = &proxy.Auth{User: conf.ProxyUser, Password: conf.ProxyPassword}
		}
		d, err := proxy.SOCKS5("tcp", conf.ProxyAddress, auth, nil)
		if err != nil {
			return nil, fmt.Errorf("无法创建代理连接器: %v", err)
		}
		kConf.Net.Proxy.Dialer = d
	}

	kConf.Producer.MaxMessageBytes = conf.MaxMessageBytes
	switch strings.ToLower(conf.Compression) {
	case "gzip":
		kConf.Producer.Compression = sarama.CompressionGZIP
	case "snappy":
		kConf.Producer.Compression = sarama.CompressionSnappy
	case "lz4":
		kConf.Producer.Compression = sarama.CompressionLZ4
	case "zstd":
		kConf.Producer.Compression = sarama.CompressionZSTD
	default:
		kConf.Producer.Compression = sarama.CompressionNone
	}
	kConf.Producer.CompressionLevel = conf.CompressionLevel
	kConf.Producer.Return.Successes = conf.EnabledSuccessesChannel // 如果启用了该选项，交付成功的消息将在Successes通道上返回
	kConf.Producer.Return.Errors = conf.EnabledErrorsChannel       // 如果启用了该选项，未交付的消息将在Errors通道上返回，包括error

	kConf.Producer.Retry.Max = conf.RetryCount
	kConf.Producer.Retry.Backoff = time.Duration(conf.RetryInterval) * time.Millisecond

	kConf.ChannelBufferSize = conf.ChannelBufferSize
	if strings.HasPrefix(conf.KafkaVersion, "v") {
		conf.KafkaVersion = conf.KafkaVersion[1:]
	}
	kConf.Version, err = sarama.ParseKafkaVersion(conf.KafkaVersion)
	if err != nil {
		return nil, fmt.Errorf("无法解析版本号: %v", err)
	}

	conf.kConf = kConf
	return conf, nil
}

// 生成同步客户端
func (k *KafkaProducer) makeSyncClient(name string) (conn.IInstance, error) {
	conf, err := k.makeConf(name)
	if err != nil {
		return nil, err
	}

	// 同步生产者必须设置这两个为true
	conf.kConf.Producer.Return.Successes = true
	conf.kConf.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(strings.Split(conf.Address, ","), conf.kConf)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %s", err)
	}
	return &syncInstance{syncProducer: producer}, nil
}

// 生成异步客户端
func (k *KafkaProducer) makeAsyncClient(name string) (conn.IInstance, error) {
	conf, err := k.makeConf(name)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewAsyncProducer(strings.Split(conf.Address, ","), conf.kConf)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %s", err)
	}
	return &asyncInstance{asyncProducer: producer}, nil
}

func (k KafkaProducer) Close() {
	k.connSync.CloseAll()
	k.connAsync.CloseAll()
}
