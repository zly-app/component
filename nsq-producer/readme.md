
# nsq生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/nsqio/go-nsq](https://github.com/nsqio/go-nsq)

```text
nsq_producer.INsqProducer                   组件接口
nsq_producer.NewNsqProducer                 创建nsq生产者组件

c.GetNsqProducer(name ...string)            获取nsq生产者
```

# 配置

> 默认组件类型为 `nsq-producer`

```toml
[components.nsq-producer.default]
Address="localhost:4150"                # 地址: localhost:4150
AuthSecret=""                           # 验证秘钥
HeartbeatInterval=30000                 # 心跳间隔(毫秒), 不能超过ReadTimeout
ReadTimeout=30000                       # 超时(毫秒
WriteTimeout=5000                       # 超时(毫秒
DialTimeout=2000                        # 超时(毫秒
```
