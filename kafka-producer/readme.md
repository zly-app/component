# kafka生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/Shopify/sarama](https://github.com/Shopify/sarama)

```text
kafka_producer.IKafkaProducer               组件接口
kafka_producer.NewKafkaProducer             创建kafka生产者组件

c.GetKafkaSyncProducer(name ...string)      获取kafka同步生产者
c.GetKafkaAsyncProducer(name ...string)     获取kafka异步生产者
```

# 同步生产者示例

```go
package main

import (
	"strconv"
	"go.uber.org/zap"
	"github.com/zly-app/component/kafka-producer"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
)

type Component struct {
	core.IComponent
	kafka_producer.IKafkaProducer
}

func (c *Component) Close() {
	c.IComponent.Close()
	c.IKafkaProducer.Close()
}

func main() {
	app := zapp.NewApp("test",
		zapp.WithCustomComponent(func(app core.IApp) core.IComponent {
			return &Component{
				IComponent:     app.GetComponent(),
				IKafkaProducer: kafka_producer.NewKafkaProducer(app),
			}
		}),
	)
	defer app.Exit()

	c := app.GetComponent().(*Component)
	p := c.GetKafkaSyncProducer()
	for i := 0; i < 10; i++ {
		msg := &kafka_producer.ProducerMessage{
			Topic:    "test",
			Key:      kafka_producer.StringEncoder("k" + strconv.Itoa(i)),
			Value:    kafka_producer.StringEncoder("v" + strconv.Itoa(i)),
			Metadata: i,
		}
		partition, offset, err := p.SendMessage(msg)
		if err != nil {
			c.Error("发送失败", zap.Error(err))
			continue
		}
		c.Info("发送成功", zap.Int32("partition", partition), zap.Int64("offset", offset))
	}
}
```

# 异步生产者示例

```go
package main

import (
	"time"
	"strconv"
	"go.uber.org/zap"
	"github.com/zly-app/component/kafka-producer"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
)

type Component struct {
	core.IComponent
	kafka_producer.IKafkaProducer
}

func (c *Component) Close() {
	c.IComponent.Close()
	c.IKafkaProducer.Close()
}

func main() {
	app := zapp.NewApp("test",
		zapp.WithCustomComponent(func(app core.IApp) core.IComponent {
			return &Component{
				IComponent:     app.GetComponent(),
				IKafkaProducer: kafka_producer.NewKafkaProducer(app),
			}
		}),
	)
	defer app.Exit()

	c := app.GetComponent().(*Component)
	p := c.GetKafkaAsyncProducer()
	for i := 0; i < 10; i++ {
		p.Input() <- &kafka_producer.ProducerMessage{
			Topic:    "test",
			Key:      kafka_producer.StringEncoder("k" + strconv.Itoa(i)),
			Value:    kafka_producer.StringEncoder("v" + strconv.Itoa(i)),
			Metadata: i,
		}
	}

	ta := time.NewTimer(time.Second * 3)
	defer ta.Stop()
	for i := 0; i < 10; i++ {
		select {
		case msg := <-p.Errors():
			c.Error("消息发送失败", zap.Any("msg", msg.Msg), zap.Error(msg.Err))
		case msg := <-p.Successes():
			c.Info("消息发送成功", zap.Any("msg", msg))
		case <-ta.C:
			c.Error("超时")
			return
		}
	}
}
```

# 配置

> 默认组件类型为 `kafka-producer`

```toml
[components.kafka-producer.default]
Address = "localhost:9092"                # 地址, 多个地址用半角逗号连接
```
