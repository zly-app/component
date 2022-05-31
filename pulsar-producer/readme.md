
# pulsar生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/apache/pulsar-client-go](https://github.com/apache/pulsar-client-go)

# 示例

1. 添加配置文件 `configs/default.yml`. 更多配置参考[这里](./config.go)

```yaml
components:
  pulsar-producer:
    default:
      url: pulsar://localhost:6650
      topic: persistent://public/default/test
```

2. 添加代码

```go
package main

import (
	"strconv"

	"go.uber.org/zap"

	"github.com/zly-app/component/pulsar-producer"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test")
	defer app.Exit()

	creator := pulsar_producer.NewProducerCreator(app) // 创建一个建造者

	for i := 0; i < 10; i++ {
		producer := creator.GetPulsarProducer() // 获取生产者
		msg := &pulsar_producer.ProducerMessage{Payload: []byte("hello" + strconv.Itoa(i))}
		_, err := producer.Send(app.BaseContext(), msg) // 发送同步消息
		if err != nil {
			app.Error("发送失败", zap.Error(err))
			return
		}
		app.Info("发送成功")
	}
}
```
