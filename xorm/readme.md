# xorm组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [xorm.io/xorm](https://gitea.com/xorm/xorm)

```go
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()

    creator := xorm.NewXormCreator(app) // 创建建造者

    client := creator.GetXorm("default") // 通过建造者获取客户端
    client.Close()
}
``` 

# 配置

> 默认组件类型为 `xorm`, 完整配置说明参考 [Config](./config.go)

+ mysql

```yaml
components:
  xorm:
    default:
      Driver: mysql # 驱动, 支持 mysql, postgres, sqlite3, mssql
      Source: 'user:passwd@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local' # 连接源
      MaxIdleConns: 2 # 最大空闲连接数
      MaxOpenConns: 5 # 最大连接池个数
      ConnMaxLifetimeSec: 0 # 最大续航时间, 秒, 0表示无限
      TableMapperRule: 'GonicMapper' # 表映射规则, 可选 SnakeMapper, SameMapper, GonicMapper, 默认为 GonicMapper
      ColumnMapperRule: 'GonicMapper' # 列映射规则, 可选 SnakeMapper, SameMapper, GonicMapper, 默认为 GonicMapper
      DisableOpenTrace: false # 关闭开放链路追踪
      TZ: Asia/Shanghai # 时区
```

+ sqllite3

```yaml
components:
  xorm:
    default:
      Driver: sqllite3 # 驱动
      Source: test.db # 连接源
```

# 驱动支持

+ `mysql`
+ `postgres`
+ `sqlite3`
+ `mssql`

# 链路追踪

使用 [github.com/opentracing/opentracing-go](https://github.com/opentracing/opentracing-go) 作为链路追踪框架

## 直接使用

每一次sql操作都会自动创建独立的trace记录

## 在函数中使用, 作为子span

1. 将engine的某一次查询作为子span记录

```go
func MyFun(c xorm.IXorm){
    span := opentracing.StartSpan("my_fun") // 创建span
    defer span.Finish() // 别忘记关闭
    ctx := opentracing.ContextWithSpan(context.Background(), span) // 将span存入ctx

    var a interface{}
    c.GetXorm().Context(ctx). // 设置ctx, 会根据ctx中带的span自动生成子span
        Sql(`select 1;`).Find(&a)
}
```

2. 将session中的每一次查询作为子span记录

```go
func MyFun(c xorm.IXorm){
    span := opentracing.StartSpan("my_fun") // 创建span
    defer span.Finish() // 别忘记关闭
    ctx := opentracing.ContextWithSpan(context.Background(), span) // 将span存入ctx
    session := c.GetXorm().NewSession().Context(ctx) // 设置ctx, session中的每次操作都会自动生成一个子span

    var a, b interface{}
    session.Sql(`select 1;`).Find(&a)
    session.Sql(`select 1;`).Find(&b)
}
```
