# sqlx组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)

```text
sqlx.ISqlx                  组件接口
sqlx.NewSqlx                创建组件

c.GetSqlx(name ...string)           获取sqlx.DB
``` 

# 配置

> 默认组件类型为 `sqlx`, 完整配置说明参考 [Config](./config.go)

```toml
[components.sqlx.default]
# 驱动
Driver = "sqllite3"
# 连接源
Source = "test.db"
```

# 驱动支持

+ `clickhouse`
+ `mysql`
+ `postgres`
+ `sqlite3`
+ `mssql`
