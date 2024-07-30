
# mqtt生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang)

# 示例

1. 添加配置文件 `configs/default.yml`. 更多配置参考[这里](./config.go)

```yaml
components:
  mqtt-producer:
    default:
      Server: localhost:1883 # mqtt服务地址
      WaitConnectedTimeMs: 5000 # 等待连接超时时间, 单位毫秒
      
      User: '' # mqtt用户名
      Password: '' # mqtt密码
```

2. 添加代码

```go
package main

import (
	"strconv"

	"go.uber.org/zap"

	mqtt_producer "github.com/zly-app/component/mqtt-producer"
	"github.com/zly-app/zapp"
)

func main() {
	app := zapp.NewApp("test-pulsar")
	defer app.Exit()

	for i := 0; i < 3; i++ {
		producer := mqtt_producer.GetDefClient() // 获取默认生产者客户端
		msg := &mqtt_producer.ProducerMessage{Topic: "test", Qos: 1, Payload: "hello" + strconv.Itoa(i)}
		err := producer.Send(app.BaseContext(), msg) // 发送同步消息
		if err != nil {
			app.Error("发送失败", zap.Error(err))
			return
		}
		app.Info("发送成功")
	}
}
```
