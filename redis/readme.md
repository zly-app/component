
# redis客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/go-redis/redis/v8](https://github.com/go-redis/redis)

```text
redis.IRedis                组件接口
redis.NewRedis              创建redis组件

c.GetRedis(name ...string)          获取redis客户端
```

# 配置

> 默认组件类型为 `redis`

```toml
[components.redis.default]
Address="localhost:6379"                # 地址: host1:port1,host2:port2
UserName=""                             # 用户名                     
Password="yourpwd"                      # 密码
DB=0                                    # db, 只有非集群有效
IsCluster=false                         # 是否为集群
MinIdleConns=3                          # 最小空闲连接数
PoolSize=10                             # 客户端池大小
ReadTimeout=5000                        # 超时(毫秒
WriteTimeout=5000                       # 超时(毫秒
DialTimeout=5000                        # 超时(毫秒
```

# 链路追踪

使用 [github.com/opentracing/opentracing-go](https://github.com/opentracing/opentracing-go) 作为链路追踪框架

## 直接使用

每一次操作都会自动创建独立的trace记录

## 在函数中使用, 作为子span

1. 将某一次操作作为子span记录

```go
func MyFun(c redis.IRedis){
    span := opentracing.StartSpan("my_fun") // 创建span
    defer span.Finish() // 别忘记关闭
    ctx := opentracing.ContextWithSpan(context.Background(), span) // 将span存入ctx

    var a interface{}
    _, _ = c.GetRedis().Get(ctx, "a").Result() // 设置ctx, 会根据ctx中带的span自动生成子span
}
```

2. 将pipeline中的所有操作作为子span记录

```go
func MyFun(c redis.IRedis){
    span := opentracing.StartSpan("my_fun") // 创建span
    defer span.Finish() // 别忘记关闭
    ctx := opentracing.ContextWithSpan(context.Background(), span) // 将span存入ctx

    var a interface{}
    _, _ = c.GetRedis().Pipelined(ctx, // 设置ctx, pipeline中的每个操作都会自动生成一个子span
        func(pipe redis.Pipeliner) error {
            return nil
        })
}
```
