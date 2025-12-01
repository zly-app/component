# kafka生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/Shopify/sarama](https://github.com/Shopify/sarama)

```text
kafka_producer.GetDefClient()          // 获取kafka同步生产者
kafka_producer.GetDefAsyncClient()     // 获取kafka异步生产者
```

# 同步生产者示例

```go
package main

import (
	"strconv"
	"go.uber.org/zap"
	"github.com/zly-app/component/kafka-producer"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test")
	defer app.Exit()

	for i := 0; i < 10; i++ {
		msg := &kafka_producer.ProducerMessage{
			Topic:    "test",
			Key:      kafka_producer.StringEncoder("k" + strconv.Itoa(i)),
			Value:    kafka_producer.StringEncoder("v" + strconv.Itoa(i)),
			Metadata: i,
		}
		partition, offset, err := kafka_producer.GetDefClient().SendMessage(msg)
		if err != nil {
			app.Error("发送失败", zap.Error(err))
			continue
		}
		app.Info("发送成功", zap.Int32("partition", partition), zap.Int64("offset", offset))
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
)

func main() {
	app := zapp.NewApp("test")
	defer app.Exit()

	for i := 0; i < 10; i++ {
		client, _ := kafka_producer.GetDefAsyncClient()
		client.Input() <- &kafka_producer.ProducerMessage{
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
			app.Error("消息发送失败", zap.Any("msg", msg.Msg), zap.Error(msg.Err))
		case msg := <-p.Successes():
			app.Info("消息发送成功", zap.Any("msg", msg))
		case <-ta.C:
			app.Error("超时")
			return
		}
	}
}
```

# 配置

> 默认组件类型为 `kafka-producer`

```yaml
components:
  kafka-producer:
    default:
      Address: "localhost:9092"                # 地址, 多个地址用半角逗号连接
```
