
# grpc客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

```text
grpc_client.IGrpcClient                 组件接口
grpc_client.NewGrpcClient               创建grpc客户端组件

c.RegistryGrpcClientCreator             注册grpc客户端创造者
c.GetGrpcClient(name ...string)         获取grpc客户端
```

# 示例

+ [grpc服务端](https://github.com/zly-app/service/tree/master/grpc/example/grpc-s)
+ [grpc客户端](./example/grpc-c)

# 配置

> 默认组件类型为 `grpc-client`

```toml
[components.grpc-client.default]
# 链接地址
Address="127.0.0.1:3000"
# 注册器, 默认为 local
Registry="local"
# 负载均衡, 默认为 round_robin
Balance="round_robin"
# 连接超时(毫秒), 默认为 5000
DialTimeout=5000
```
