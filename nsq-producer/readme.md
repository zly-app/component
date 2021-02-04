
# nsq生产者组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

```text
nsq_producer.INsqProducerComponent          组件接口
nsq_producer.NewNsqProducer                 创建nsq生产者组件

c.GetNsqProducer(name ...string)            获取nsq生产者
```

# 配置

> 默认组件类型为 `nsq-producer`

```toml
[components.nsq-producer.default]
Address="localhost:4150"                # 地址: localhost:4150
AuthSecret=""                           # 验证秘钥
ReadTimeout=5000                        # 超时(毫秒
WriteTimeout=5000                       # 超时(毫秒
DialTimeout=5000                        # 超时(毫秒
```
