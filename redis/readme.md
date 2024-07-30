
# redis客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/redis/go-redis/v9](https://github.com/go-redis/redis)

```go
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()
    client := creator.GetDefClient() // 获取默认客户端
}
```

# 配置

> 默认组件类型为 `redis`

```yaml
components:
  redis:
    default:
      Address: localhost:6379 # 地址: host1:port1,host2:port2
      UserName: "" # 用户名                     
      Password: "" # 密码
      DB: 0 # db, 只有非集群有效

      MinIdle: 2 # 最小闲置连接数
      MaxIdle: 4 # 最大闲置连接数
      PoolSize: 10 # 客户端池大小
      IdleTimeout: 3600 # 空闲链接超时时间, 单位秒, 如果一个连接长时间未使用将被视为连接无效, 小于1表示永不超时
      WaitTimeout: 5 # 等待获取连接的超时时间, 单位秒
      ConnectTimeout: 5 # 连接超时, 单位秒
      MaxConnLifetime: 3600 # 一个连接最大存活时间, 单位秒, 小于1表示不限制

      MaxRetries: 0 # 操作尝试次数, <1 表示不重试
      ReadTimeoutSec: 5 # 超时, 秒
      WriteTimeoutSec: 5 # 超时, 秒
```
