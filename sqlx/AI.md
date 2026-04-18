# SQLx 组件 - AI 参考文档

## 概述

基于 [jmoiron/sqlx](https://github.com/jmoiron/sqlx) 的 SQL 数据库客户端组件，用于 zapp 框架。提供类型安全的数据库操作，自动集成 filter 链、事务管理和连接池。支持 MySQL、PostgreSQL、SQLite3、MSSQL 和 ClickHouse 驱动。

## 包路径

`github.com/zly-app/component/sqlx`

## 核心类型

### `Client`（接口）

顶层数据库客户端。所有操作经过 zapp filter 链，支持链路追踪/日志/指标采集。

| 方法 | 签名 | 说明 |
|------|------|------|
| `GetDB` | `() *sqlx.DB` | 获取底层 sqlx.DB 实例 |
| `Unsafe` | `() Client` | 返回不检查结构体字段数量的客户端 |
| `Find` | `(ctx context.Context, dest interface{}, query string, args ...interface{}) error` | 查询多行到结构体/基础类型切片 |
| `FindOne` | `(ctx context.Context, dest interface{}, query string, args ...interface{}) error` | 查询单行到结构体/基础类型 |
| `FindColumn` | `(ctx context.Context, dest []interface{}, query string, args ...interface{}) error` | 查询单行多列到独立变量 |
| `Exec` | `(ctx context.Context, query string, args ...interface{}) (sql.Result, error)` | 执行写操作（INSERT/UPDATE/DELETE） |
| `Query` | `(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)` | 原始查询，返回 sql.Rows |
| `Transaction` | `(ctx context.Context, f TxFunc) error` | 在标准 sql.Tx 事务中执行函数。返回 nil 自动提交，返回 error 自动回滚 |
| `TransactionX` | `(ctx context.Context, f TxxFunc) error` | 在 sqlx.Tx 增强事务中执行函数（支持 Find/FindOne）。提交/回滚规则同 Transaction |

### `Tx`（接口）

标准事务上下文（`sql.Tx`），在 `Transaction()` 回调中使用。

| 方法 | 签名 | 说明 |
|------|------|------|
| `Tx` | `() *sql.Tx` | 获取底层 sql.Tx |
| `FindColumn` | `(ctx context.Context, dest []interface{}, query string, args ...interface{}) error` | 查询单行多列 |
| `FindToStructs` | `(ctx context.Context, dest interface{}, query string, args ...interface{}) error` | 查询多行到结构体 |
| `Exec` | `(ctx context.Context, query string, args ...interface{}) (sql.Result, error)` | 执行写操作 |
| `Query` | `(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)` | 原始查询 |

### `Txx`（接口）

sqlx 增强事务，在 `TransactionX()` 回调中使用。

| 方法 | 签名 | 说明 |
|------|------|------|
| `Tx` | `() *sql.Tx` | 获取底层 sql.Tx |
| `Txx` | `() *sqlx.Tx` | 获取底层 sqlx.Tx |
| `Unsafe` | `() Txx` | 返回不检查字段数的 Txx |
| `Find` | `(ctx context.Context, dest interface{}, query string, args ...interface{}) error` | 查询多行 |
| `FindOne` | `(ctx context.Context, dest interface{}, query string, args ...interface{}) error` | 查询单行 |
| `FindColumn` | `(ctx context.Context, dest []interface{}, query string, args ...interface{}) error` | 查询单行多列 |
| `Exec` | `(ctx context.Context, query string, args ...interface{}) (sql.Result, error)` | 执行写操作 |
| `Query` | `(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)` | 原始查询 |

### `SqlxConfig`（结构体）

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Driver` | `string` | 必填 | 数据库驱动：mysql、postgres、sqlite3、mssql、clickhouse |
| `Source` | `string` | 必填 | 连接源字符串（DSN） |
| `MaxIdleConns` | `int` | 2 | 最大空闲连接数 |
| `MaxOpenConns` | `int` | 5 | 最大打开连接数 |
| `ConnMaxLifetime` | `int` | 0 | 连接最大存活时间（毫秒），0 表示不限制 |

### `Creator`（接口）

| 方法 | 签名 | 说明 |
|------|------|------|
| `GetClient` | `(name string) (Client, error)` | 获取指定名称的客户端 |
| `GetDefClient` | `() Client` | 获取默认客户端（不返回 error，失败时返回 errClient） |

### 事务选项函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `WithTxIsolation` | `(level sql.IsolationLevel) TxOption` | 设置事务隔离级别 |
| `WithTxReadOnly` | `() TxOption` | 设置事务为只读 |

### 工具函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `GetModelSelectField` | `(model interface{}) string` | 通过反射从结构体 `db` tag 提取 SELECT 字段列表 |
| `GetModelSelectFieldByTagName` | `(model interface{}, tagName string) string` | 使用自定义 tag 名提取 SELECT 字段 |

## 包级便捷函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `GetClient` | `(name string) (Client, error)` | 通过默认 Creator 获取指定名称的客户端 |
| `GetDefClient` | `() Client` | 通过默认 Creator 获取默认客户端 |

## 错误处理

客户端创建失败时，`GetDefClient()` 返回 `errClient`（空对象模式）而非 nil。`errClient` 的所有方法都返回创建时的原始错误，避免空指针异常。

## YAML 配置

组件类型：`sqlx`

```yaml
components:
  sqlx:
    default:
      Driver: mysql
      Source: 'user:passwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local'
      MaxIdleConns: 2
      MaxOpenConns: 5
      ConnMaxLifetimeSec: 0
```

## 使用模式

### 基本查询

```go
client := sqlx.GetDefClient()

// 查询到结构体切片
var users []User
err := client.Find(ctx, &users, "SELECT * FROM users WHERE age > ?", 18)

// 查询单个结构体
var user User
err := client.FindOne(ctx, &user, "SELECT * FROM users WHERE id = ?", 1)

// 查询到基础类型切片
var ids []int
err := client.Find(ctx, &ids, "SELECT id FROM users")
```

### 写操作

```go
result, err := client.Exec(ctx, "INSERT INTO users (name, age) VALUES (?, ?)", "Alice", 30)
```

### 标准事务

```go
err := client.Transaction(ctx, func(ctx context.Context, tx sqlx.Tx) error {
    var a int
    var b string
    _ = tx.FindColumn(ctx, []interface{}{&a, &b}, "SELECT a, b FROM test WHERE id = ?", 1)
    _, err := tx.Exec(ctx, "UPDATE test SET a = ? WHERE id = ?", a+1, 1)
    return err // 返回 nil 自动提交，返回 error 自动回滚
})
```

### sqlx 增强事务

```go
err := client.TransactionX(ctx, func(ctx context.Context, txx sqlx.Txx) error {
    var user User
    _ = txx.FindOne(ctx, &user, "SELECT * FROM users WHERE id = ?", 1)
    return nil
})
```

## Filter 链集成

所有数据库操作经过 zapp filter 链（`filter.GetClientFilter()`），组件类型为 `"sqlx"`，方法名为操作名（Find/FindOne/Exec/Query/Transaction/TransactionX）。可支持：
- OpenTelemetry 链路追踪传播
- 请求/响应日志记录
- 指标采集
- 自定义中间件注入
