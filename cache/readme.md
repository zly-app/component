
# cache组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/zlyuancn/zcache](https://github.com/zlyuancn/zcache)

```text
cache.ICache            组件接口
cache.NewCache          创建cache组件

c.GetCache(name ...string)          获取cache
```

# 示例

```go
package main

import (
	"github.com/zlyuancn/zcache"
	"go.uber.org/zap"

	"github.com/zly-app/component/cache"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
)

type Component struct {
	core.IComponent
	cache.ICache
}

func main() {
	app := zapp.NewApp("test", zapp.WithCustomComponent(func(app core.IApp) core.IComponent {
		return &Component{
			IComponent: app.GetComponent(),
			ICache:     cache.NewCache(app),
		}
	}))

	c := app.GetComponent().(*Component).GetCache()

	err := c.Save("test", "hello", 0, zcache.QC().Args("k1"))
	app.Info("保存", zap.Error(err))

	var s string
	err = c.Query("test", &s, zcache.QC().Args("k1"))
	app.Info("获取", zap.String("s", s), zap.Error(err))
}
```

# 配置

> 默认组件类型为 `cache`

```toml
[components.cache.default]
# 缓存db; default, no, memory, redis
CacheDB="memory"
# 编解码器; default, byte, json, jsoniter, msgpack, proto_buffer
Codec="msgpack"
# 在缓存故障时直接返回缓存错误
DirectReturnOnCacheFault=true
# 注册加载器时如果加载器已存在会panic, 设为false会替换旧的加载器
PanicOnLoaderExists=true
# 单跑; default, no, single
SingleFlight="single"
# 默认有效时间, 毫秒, <= 0 表示永久
DefaultExpire=0
# 默认最大有效时间, 毫秒, 如果 > 0 且 DefaultExpire > 0, 则默认有效时间在 [DefaultExpire, DefaultExpireMax-1] 区间随机
DefaultExpireMax=0
```

## 使用内存时添加以下配置

```toml
[components.cache.default]
# ...
# 清除过期key时间间隔(毫秒)
MemoryCacheDB.CleanupInterva=300000
```

## 使用redis时添加一下配置

```toml
[components.cache.default]
# ...
# key前缀
RedisCacheDB.KeyPrefix=""
# 地址: host1:port1,host2:port2
RedisCacheDB.Address="localhost:6379"
# 密码
RedisCacheDB.Password=""
# db, 只有单点有效
RedisCacheDB.DB=0
# 是否为集群
RedisCacheDB.IsCluster=false
# 客户端池大小
RedisCacheDB.PoolSize=10
# 读取超时(毫秒
RedisCacheDB.ReadTimeout=5000
# 写入超时(毫秒
RedisCacheDB.WriteTimeout=5000
# 连接超时(毫秒
RedisCacheDB.DialTimeout=5000
```