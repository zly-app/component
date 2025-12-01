
# nsq生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/nsqio/go-nsq](https://github.com/nsqio/go-nsq)

```text
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()
    client, _ := nsq_producer.GetDefClient() // 获取默认生产者客户端
}
```

# 配置

> 默认组件类型为 `nsq-producer`

```yaml
components:
  nsq-producer:
    default:
      Address: "localhost:4150"                # 地址: localhost:4150
      AuthSecret: ""                           # 验证秘钥
      HeartbeatInterval: 30000                 # 心跳间隔(毫秒), 不能超过ReadTimeout
      ReadTimeout: 30000                       # 超时(毫秒
      WriteTimeout: 5000                       # 超时(毫秒
      DialTimeout: 2000                        # 超时(毫秒
```
