
# es7客户端组件插件

> 提供用于 https://github.com/zly-app/zapp 的组件插件

# 说明

```text
es7.IES7Component           组件接口
es7.NewES7                  创建es7组件

c.GetES7(name ...string)            获取es7客户端
```

# 配置

> 默认组件类型为 `es7`

```toml
[components.es7.default]
Address="http://localhost:9200"     # 地址: http://localhost1:9200,http://localhost2:9200
UserName=""                         # 用户名
Password=""                         # 密码
DialTimeout=5000                    # 连接超时(毫秒
Sniff=false                         # 开启嗅探器
Healthcheck=true                    # 心跳检查(默认true
Retry=0                             # 重试次数
RetryInterval=0                     # 重试间隔(毫秒)
GZip=false                          # 启用gzip压缩
```
