
# ssdb客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/seefan/gossdb](https://github.com/seefan/gossdb)

```text
ssdb.ISsdb                  组件接口
ssdb.NewSsdb                创建ssdb组件

c.GetSsdb(name ...string)   获取客户端, 使用完后记得关闭
```

# 配置

> 默认组件类型为 `ssdb`

```toml
[components.ssdb.default]
# host
Host="localhost"
# port
Port=8888
# 密码
Password=""
# 获取客户端超时(毫秒)
GetClientTimeout=5000
# 获取客户端超时等待队列大小, 当连接池满后, 新建连接将等待池中连接释放后才可以继续, 超出队列的会抛出异常
GetClientWaitQueueSize=1000
# 读取超时(毫秒
ReadTimeout=5000
# 写入超时(毫秒
WriteTimeout=5000
# 连接超时(毫秒
DialTimeout=5000
# 最小连接池数
MinPoolSize=20
# 最大连接池个数
MaxPoolSize=100
# 是否启用重试, 设置为true时, 如果请求失败会再重试一次
RetryEnabled=false
```
