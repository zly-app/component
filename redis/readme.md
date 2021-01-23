
# redis客户端组件插件

> 提供用于 https://github.com/zly-app/zapp 的组件插件

# 说明

```text
redis.IRedisComponent       组件接口
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
