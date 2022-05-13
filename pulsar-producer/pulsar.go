package pulsar_producer

import (
	"fmt"

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
	err := p.app.GetConfig().ParseComponentConfig(DefaultComponentType, name, conf, false)
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
		app: app,
	}
	return p
}

var _ IPulsarProducer = (*PulsarProducer)(nil)
var _ conn.IInstance = (*PulsarProducer)(nil)

type PulsarProducer struct {
	client   pulsar.Client
	producer pulsar.Producer
}

func (p *PulsarProducer) Close() {
	p.producer.Close()
	p.client.Close()
}

// 创建生产者
func NewProducer(conf *Config) (*PulsarProducer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		Logger:       log.DefaultNopLogger(),
		URL:          "pulsar://localhost:6600,localhost:6650",
		ListenerName: "external",
	})
	if err != nil {
		return nil, fmt.Errorf("创建pulsar客户端失败: %v", err)
	}

	po := pulsar.ProducerOptions{}
	producer, err := client.CreateProducer(po)
	if err != nil {
		return nil, fmt.Errorf("创建pulsar生产者失败: %v", err)
	}

	return &PulsarProducer{
		client:   client,
		producer: producer,
	}, nil
}
