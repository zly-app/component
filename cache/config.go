/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/25
   Description :
-------------------------------------------------
*/

package cache

import (
	"github.com/zly-app/zapp/core"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "cache"

const (
	// 默认缓存db
	defaultCacheDB = "memory"
	// 默认编解码器
	defaultCodec = "msgpack"
	// 默认在缓存故障时直接返回缓存错误
	defaultDirectReturnOnCacheFault = true
	// 默认注册加载器时如果加载器已存在会panic
	defaultPanicOnLoaderExists = true
	// 默认单跑
	defaultSingleFlight = "single"

	// 默认清除过期key时间间隔
	defaultMemoryCacheDBCleanupInterval = 300000

	// 默认客户端池大小
	defaultRedisCacheDBPoolSize = 10
	// 默认读取超时(毫秒
	defaultRedisCacheDBReadTimeout = 5000
	// 默认写入超时(毫秒
	defaultRedisCacheDBWriteTimeout = 5000
	// 默认连接超时(毫秒
	defaultRedisCacheDBDialTimeout = 5000
	// 默认操作超时(毫秒
	defaultRedisCacheDBDoTimeout = 5000
)

// 缓存配置
type CacheConfig struct {
	CacheDB                  string // 缓存db; default, no, memory, redis
	Codec                    string // 编解码器; default, byte, json, jsoniter, msgpack, proto_buffer
	DirectReturnOnCacheFault *bool  // 在缓存故障时直接返回缓存错误
	PanicOnLoaderExists      *bool  // 注册加载器时如果加载器已存在会panic, 设为false会替换旧的加载器
	SingleFlight             string // 单跑; default, no, single
	DefaultExpire            int64  // 默认有效时间, 毫秒, <= 0 表示永久
	DefaultExpireMax         int64  // 默认最大有效时间, 毫秒, 如果 > 0 且 DefaultExpire > 0, 则默认有效时间在 [DefaultExpire, DefaultExpireMax-1] 区间随机

	MemoryCacheDB struct {
		CleanupInterval int64 // 清除过期key时间间隔(毫秒)
	}
	RedisCacheDB struct {
		KeyPrefix    string // key前缀
		Address      string // 地址: host1:port1,host2:port2
		Password     string // 密码
		DB           int    // db, 只有单点有效
		IsCluster    bool   // 是否为集群
		PoolSize     int    // 客户端池大小
		ReadTimeout  int64  // 读取超时(毫秒
		WriteTimeout int64  // 写入超时(毫秒
		DialTimeout  int64  // 连接超时(毫秒
		DoTimeout    int64  // 操作超时
	}
}

func (conf *CacheConfig) Check() {
	if conf.CacheDB == "default" || conf.CacheDB == "" {
		conf.CacheDB = defaultCacheDB
	}
	if conf.Codec == "default" || conf.Codec == "" {
		conf.Codec = defaultCodec
	}
	if conf.DirectReturnOnCacheFault == nil {
		v := defaultDirectReturnOnCacheFault
		conf.DirectReturnOnCacheFault = &v
	}
	if conf.PanicOnLoaderExists == nil {
		v := defaultPanicOnLoaderExists
		conf.PanicOnLoaderExists = &v
	}
	if conf.SingleFlight == "default" || conf.SingleFlight == "" {
		conf.SingleFlight = defaultSingleFlight
	}

	if conf.MemoryCacheDB.CleanupInterval < 1 {
		conf.MemoryCacheDB.CleanupInterval = defaultMemoryCacheDBCleanupInterval
	}

	if conf.RedisCacheDB.PoolSize < 1 {
		conf.RedisCacheDB.PoolSize = defaultRedisCacheDBPoolSize
	}
	if conf.RedisCacheDB.ReadTimeout < 1 {
		conf.RedisCacheDB.ReadTimeout = defaultRedisCacheDBReadTimeout
	}
	if conf.RedisCacheDB.WriteTimeout < 1 {
		conf.RedisCacheDB.WriteTimeout = defaultRedisCacheDBWriteTimeout
	}
	if conf.RedisCacheDB.DialTimeout < 1 {
		conf.RedisCacheDB.DialTimeout = defaultRedisCacheDBDialTimeout
	}
	if conf.RedisCacheDB.DoTimeout < 1 {
		conf.RedisCacheDB.DoTimeout = defaultRedisCacheDBDoTimeout
	}
}
