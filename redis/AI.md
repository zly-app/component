# Redis 组件 - AI 参考文档

## 概述

基于 [go-redis/v9](https://github.com/redis/go-redis) 的 Redis 客户端组件，用于 zapp 框架。提供 `UniversalClient` 统一支持单机、集群和 Ring 模式，内置 OpenTelemetry 链路追踪/指标采集和 zapp filter 链集成。

## 包路径

`github.com/zly-app/component/redis`

## 核心类型

### `Creator`（接口）

| 方法 | 签名 | 说明 |
|------|------|------|
| `GetClient` | `(name string) (UniversalClient, error)` | 获取指定名称的 Redis 客户端 |
| `GetDefClient` | `() (UniversalClient, error)` | 获取默认 Redis 客户端 |

### `UniversalClient`（类型别名）

`type UniversalClient = redis.UniversalClient` — 统一客户端接口，支持单机、集群和 Ring 拓扑。拥有 `redis.Cmdable` 接口的所有方法。

### `RedisConfig`（结构体）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Address` | `string` | 必填 | Redis 地址：`host1:port1,host2:port2`（集群用逗号分隔） |
| `UserName` | `string` | `""` | 认证用户名 |
| `Password` | `string` | `""` | 认证密码 |
| `TLSCAFile` | `string` | `""` | TLS CA 证书文件路径 |
| `InsecureSkipVerify` | `bool` | `false` | 跳过 TLS 证书验证 |
| `DB` | `int` | `0` | 数据库编号（仅非集群有效） |
| `MinIdle` | `int` | `2` | 最小空闲连接数 |
| `MaxIdle` | `int` | `MinIdle*2` | 最大空闲连接数（小于 MinIdle 时自动调整为 MinIdle*2） |
| `PoolSize` | `int` | `10` | 客户端池大小 |
| `IdleTimeout` | `int` | `3600` | 空闲连接超时时间（秒），小于 1 表示永不超时 |
| `WaitTimeout` | `int` | `5` | 等待获取连接的超时时间（秒） |
| `ConnectTimeout` | `int` | `5` | 连接超时时间（秒） |
| `MaxConnLifetime` | `int` | `3600` | 连接最大存活时间（秒），小于 1 表示不限制 |
| `MaxRetries` | `int` | `-1` | 最大重试次数，小于 1 表示禁用重试 |
| `ReadTimeoutSec` | `int` | `5` | 读取超时时间（秒） |
| `WriteTimeoutSec` | `int` | `5` | 写入超时时间（秒） |

### 重导出类型（来自 go-redis）

`link.go` 重新导出了以下 go-redis 类型，方便直接使用：

**常量/错误：**
- `redis.Nil` — Redis 空响应（键不存在）
- `redis.KeepTTL` — 保留已有 TTL
- `redis.TxFailedErr` — 事务失败错误
- `redis.ErrClosed` — 客户端已关闭错误

**命令类型：**
- `Pipeliner`、`Cmdable`、`Cmder`
- `Z`（有序集合成员）
- `*Cmd`、`BoolCmd`、`IntCmd`、`StringCmd`、`FloatCmd`、`StringSliceCmd`、`BoolSliceCmd`、`IntSliceCmd`、`FloatSliceCmd`、`DurationCmd`、`TimeCmd`、`ScanCmd`、`StringStringMapCmd`、`StringIntMapCmd`、`IntIntMapCmd`、`XMessageSliceCmd`、`XStreamSliceCmd`、`ZSliceCmd`、`ZWithKeyCmd`、`MappingCmd`、`CommandsSliceCmd`、`ClusterSlotsCmd`、`GeoLocationCmd`、`GeoSearchLocationCmd`、`GeoSearchCountCmd`、`RankWithScoreCmd`

**函数：**
- `redis.NewClient()`、`redis.NewClusterClient()`、`redis.NewFailoverClient()`、`redis.NewRing()`、`redis.NewSentinelClient()`

## 包级便捷函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `GetClient` | `(name string) (UniversalClient, error)` | 通过默认 Creator 获取指定名称的客户端 |
| `GetDefClient` | `() (UniversalClient, error)` | 通过默认 Creator 获取默认客户端 |

## 链路追踪与可观测性

### 内置 OpenTelemetry

`NewClient()` 自动启用：
- `InstrumentTracing()` — Redis 命令的 OpenTelemetry 链路追踪
- `redisotel.InstrumentMetrics()` — OpenTelemetry 指标采集

### 自定义追踪 Hook

`tracingHook` 实现 `redis.Hook` 接口，集成 zapp filter 链：

| Hook | 说明 | Filter 链数据 |
|------|------|---------------|
| `DialHook` | 拦截连接建立 | `dialReq`/`dialRsp` |
| `ProcessHook` | 拦截单条命令执行 | `cmdReq`/`cmdRsp`（记录命令名和参数） |
| `ProcessPipelineHook` | 拦截管道命令执行 | `pipelineReq`/`pipelineRsp` |

`getCmdVal()` 函数支持 20+ 种 Redis 命令返回值类型转换为字符串，用于追踪数据记录。

## YAML 配置

组件类型：`redis`

```yaml
components:
  redis:
    default:
      Address: localhost:6379
      UserName: ""
      Password: ""
      TLSCAFile: ''
      InsecureSkipVerify: false
      DB: 0
      MinIdle: 2
      MaxIdle: 4
      PoolSize: 10
      IdleTimeout: 3600
      WaitTimeout: 5
      ConnectTimeout: 5
      MaxConnLifetime: 3600
      MaxRetries: 0
      ReadTimeoutSec: 5
      WriteTimeoutSec: 5
```

## 使用模式

### 基本使用

```go
client, err := redis.GetDefClient()
if err != nil {
    // 处理错误
}

// 字符串操作
err = client.Set(ctx, "key", "value", 0).Err()
val, err := client.Get(ctx, "key").Result()

// 哈希操作
err = client.HSet(ctx, "hash", "field", "value").Err()
val, err := client.HGet(ctx, "hash", "field").Result()

// 有序集合
err = client.ZAdd(ctx, "zset", redis.Z{Score: 1, Member: "member1"}).Err()
```

### 多客户端

```go
client, err := redis.GetClient("secondary")
```

### 使用重导出类型

```go
// 使用 redis.Nil 检查键不存在
val, err := client.Get(ctx, "key").Result()
if err == redis.Nil {
    // 键不存在
}

// 使用 redis.Z 操作有序集合
members := []redis.Z{
    {Score: 1, Member: "a"},
    {Score: 2, Member: "b"},
}
client.ZAdd(ctx, "zset", members...)
```

## TLS 支持

- `TLSCAFile`：TLS 连接的 CA 证书文件路径
- `InsecureSkipVerify`：跳过服务器证书验证（用于开发/测试环境）

## 配置校验

`RedisConfig.Check()` 自动调整：
- `MaxIdle` 小于 `MinIdle` 时设为 `MinIdle * 2`
- `MaxRetries` 小于 1 时设为 `-1`（禁用重试）
