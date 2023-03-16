
# redis客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/redis/go-redis/v9](https://github.com/go-redis/redis)

```go
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()

    creator := redis.NewRedisCreator(app) // 创建redis建造者

    client := creator.GetRedis("default") // 通过redis建造者获取redis客户端
    client.Close()
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
      IsCluster: false # 是否为集群
      MinIdleConns: 2 # 最小空闲连接数
      PoolSize: 5 # 客户端池大小
      ReadTimeoutSec: 5 # 超时, 秒
      WriteTimeoutSec: 5 # 超时, 秒
      DialTimeoutSec: 5 # 超时, 秒
```
