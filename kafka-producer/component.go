package kafka_producer

import (
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"golang.org/x/net/proxy"
)

type Creator interface {
	// 获取kafka同步生产者
	GetClient(name string) Client
	// 获取kafka同步生产者
	GetDefClient() Client
	// 获取kafka异步生产者
	GetAsyncClient(name string) (AsyncClient, error)
	// 获取kafka异步生产者
	GetDefAsyncClient() (AsyncClient, error)
}

// 同步生产者
type Client interface {
	SendMessage(msg *ProducerMessage) (partition int32, offset int64, err error)
	SendMessages(msgs []*ProducerMessage) error
}

// 异步生产者
type AsyncClient interface {
	Input() chan<- *ProducerMessage
	Successes() <-chan *ProducerMessage
	Errors() <-chan *ProducerError
}

type kafkaProducer struct {
	connSync  *conn.AnyConn[Client]
	connAsync *conn.AnyConn[AsyncClient]
}

func GetCreator() Creator {
	return defCreator
}

func (k *kafkaProducer) GetClient(name string) Client {
	client, err := k.connSync.GetConn(k.makeSyncClient, name)
	if err != nil {
		return newErrClient(err)
	}
	return client
}

func (k *kafkaProducer) GetDefClient() Client {
	client, err := k.connSync.GetConn(k.makeSyncClient, consts.DefaultConfigFiles)
	if err != nil {
		return newErrClient(err)
	}
	return client
}

func (k *kafkaProducer) GetAsyncClient(name string) (AsyncClient, error) {
	client, err := k.connAsync.GetConn(k.makeAsyncClient, name)
	return client, err
}

func (k *kafkaProducer) GetDefAsyncClient() (AsyncClient, error) {
	client, err := k.connAsync.GetConn(k.makeAsyncClient, consts.DefaultComponentName)
	return client, err
}

// 生成配置
func (k *kafkaProducer) makeConf(name string) (*Config, error) {
	conf := newConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf)
	if err != nil {
		return nil, err
	}
	if err = conf.Check(); err != nil {
		return nil, fmt.Errorf("组件%s的配置错误: %s", DefaultComponentType, err)
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
func (k *kafkaProducer) makeSyncClient(name string) (Client, error) {
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
	return producer, nil
}

// 生成异步客户端
func (k *kafkaProducer) makeAsyncClient(name string) (AsyncClient, error) {
	conf, err := k.makeConf(name)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewAsyncProducer(strings.Split(conf.Address, ","), conf.kConf)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %s", err)
	}
	return producer, nil
}

func (k kafkaProducer) Close() {
	k.connSync.CloseAll()
	k.connAsync.CloseAll()
}
