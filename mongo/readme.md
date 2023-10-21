
# mongo客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [go.mongodb.org/mongo-driver](https://github.com/mongodb/mongo-go-driver)

```go
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()

    creator := mongo.NewMongoCreator(app) // 创建mongo建造者
    client := creator.GetMongo("default") // 通过mongo建造者获取mongo客户端
}
```

# 配置

> 默认组件类型为 `mongo`

```yaml
components:
  mongo:
    default:
      Address: '' # 地址: host1:port1,host2:port2
      UserName: '' # 用户名
      Password: '' # 密码
      DB: '' # db库, 如果设置了用户名则使用这个库进行认证

      MinPoolSize: 2 # 最小连接池
      MaxPoolSize: 5 # 最大连接池

      IdleTimeout: 3600  # 空闲链接超时时间, 单位秒, 如果一个连接长时间未使用将被视为连接无效, 小于1表示永不超时
      ConnectTimeout: 5  # 连接超时, 单位秒
      DoTimeoutSec: 5  # 一个操作的超时, 秒
      Driect: false # 如果设为true, 要求 address 只有一个地址, 此时会将mongo当做单节点来连接. 如果设为false, 会通过address获取其副本集/集群的其它节点地址再进行连接
```
