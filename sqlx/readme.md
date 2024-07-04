# sqlx组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)

```go
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()

    creator := sqlx.NewSqlx(app) // 创建建造者
    client := creator.GetSqlx("default") // 通过建造者获取客户端
}
```

示例转到[这里](./example)

# 配置

> 默认组件类型为 `sqlx`, 完整配置说明参考 [Config](./config.go)

+ mysql

```yaml
components:
  sqlx:
    default:
      Driver: mysql # 驱动, 支持 mysql, postgres, sqlite3, mssql
      Source: 'user:passwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local' # 连接源
      MaxIdleConns: 2 # 最大空闲连接数
      MaxOpenConns: 5 # 最大连接池个数
      ConnMaxLifetimeSec: 0 # 最大续航时间, 秒, 0表示无限
```

+ sqllite3

```yaml
components:
  sqlx:
    default:
      Driver: sqllite3 # 驱动
      Source: test.db # 连接源
```

# 驱动支持

+ `clickhouse`
+ `mysql`
+ `postgres`
+ `sqlite3`
+ `mssql`
