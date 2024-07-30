package mqtt_producer

import (
	"context"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/component/conn"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/filter"
)

type Creator interface {
	GetClient(name string) Client
	GetDefClient() Client
}

type mqttProducerCreator struct {
	conn *conn.Conn
}

func (m *mqttProducerCreator) GetClient(name string) Client {
	ins, err := m.conn.GetConn(m.makeProducer, name)
	if err != nil {
		return newErrProducer(err)
	}
	return ins.(Client)
}

func (m *mqttProducerCreator) GetDefClient() Client {
	return m.GetClient(consts.DefaultComponentName)
}

func (m *mqttProducerCreator) Close() { m.conn.CloseAll() }

func (m *mqttProducerCreator) makeProducer(name string) (conn.IInstance, error) {
	conf := NewConfig()
	err := zapp.App().GetConfig().ParseComponentConfig(DefaultComponentType, name, conf, true)
	if err != nil {
		return nil, fmt.Errorf("获取组件<%s.%s>配置失败: %v", DefaultComponentType, name, err)
	}

	producer, err := NewProducer(name, conf)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

type Client interface {
	// 发送消息
	Send(ctx context.Context, msg *ProducerMessage) error
	// 异步发送消息
	SendAsync(ctx context.Context, msg *ProducerMessage, callback func(error))
}

func GetCreator() Creator {
	return defCreator
}

type MQTTProducer struct {
	name   string
	conf   *Config
	client mqtt.Client
}

type ProducerMessage struct {
	Topic    string
	Qos      byte
	Retained bool   // 是否为保留消息
	Payload  string // 消息负载
}

func NewProducer(name string, conf *Config) (*MQTTProducer, error) {
	if err := conf.Check(); err != nil {
		return nil, fmt.Errorf("配置检查失败: %v", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(conf.Server)
	if conf.User != "" {
		opts.SetUsername(conf.User)     //账号
		opts.SetPassword(conf.Password) //密码
	}
	opts.SetConnectTimeout(time.Duration(conf.WaitConnectedTimeMs) * time.Millisecond)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	waitOk := token.WaitTimeout(time.Duration(conf.WaitConnectedTimeMs) * time.Millisecond)
	if !waitOk {
		return nil, errors.New("连接mqtt超时")
	}
	if token.Error() != nil {
		return nil, fmt.Errorf("连接mqtt失败: %s", token.Error())
	}

	return &MQTTProducer{
		name:   name,
		conf:   conf,
		client: client,
	}, nil
}

func (p *MQTTProducer) Send(ctx context.Context, msg *ProducerMessage) error {
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), p.name, "Send")
	_, err := chain.Handle(ctx, msg, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		msg := req.(*ProducerMessage)
		token := p.client.Publish(msg.Topic, msg.Qos, msg.Retained, msg.Payload)
		waitOk := token.WaitTimeout(time.Duration(p.conf.WaitConnectedTimeMs) * time.Millisecond)
		if !waitOk {
			return nil, fmt.Errorf("send to mqtt topic: %v timeout", msg.Topic)
		}
		if token.Error() != nil {
			return nil, fmt.Errorf("send to mqtt topic: %v err: %v", msg.Topic, token.Error())
		}
		return nil, err
	})
	return err
}

func (p *MQTTProducer) SendAsync(ctx context.Context, msg *ProducerMessage, callback func(error)) {
	var token mqtt.Token
	ctx, chain := filter.GetClientFilter(ctx, string(DefaultComponentType), p.name, "SendAsync")
	_, err := chain.Handle(ctx, msg, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		msg := req.(*ProducerMessage)
		token = p.client.Publish(msg.Topic, msg.Qos, msg.Retained, msg.Payload)
		return nil, nil
	})

	zapp.App().GetComponent().GetGPool().Go(func() error {
		waitOk := token.WaitTimeout(time.Duration(p.conf.WaitConnectedTimeMs) * time.Millisecond)
		if !waitOk {
			return fmt.Errorf("send to mqtt topic: %v timeout", msg.Topic)
		}
		if token.Error() != nil {
			return fmt.Errorf("send to mqtt topic: %v err: %v", msg.Topic, token.Error())
		}
		return err
	}, callback)
}

func (m *MQTTProducer) Close() {
	m.client.Disconnect(uint(m.conf.WaitConnectedTimeMs))
}
