# xorm组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [xorm.io/xorm](https://gitea.com/xorm/xorm)

```text
xorm.IXorm                  组件接口
xorm.NewXorm                创建xorm组件

c.GetXorm(name ...string)           获取xorm引擎
``` 

# 配置

> 默认组件类型为 `xorm`, 完整配置说明参考 [Config](./config.go)

```toml
[components.xorm.default]
# 驱动
Driver = "sqllite3"
# 连接源
Source = "test.db"
```

# 驱动支持

+ `mysql`
+ `postgres`
+ `sqlite3`
+ `mssql`

# 链路追踪

使用 [github.com/opentracing/opentracing-go](https://github.com/opentracing/opentracing-go) 作为链路追踪框架

## 直接使用

每一次sql操作都会自动作为独立的trace记录

## 在函数中使用, 作为子span

1. 将engine的某一次查询作为子span记录

```go
func MyFun(c xorm.IXorm){
    span := opentracing.StartSpan("my_fun") // 创建span
    defer span.Finish() // 别忘记关闭
    ctx := opentracing.ContextWithSpan(context.Background(), span) // 将span存入ctx

    var a interface{}
    c.GetXorm().Context(ctx). // 设置ctx
        Sql(`select 1;`).Find(&a)
}
```

2. 将session中的每一次查询作为子span记录

```go
func MyFun(c xorm.IXorm){
    span := opentracing.StartSpan("my_fun") // 创建span
    defer span.Finish() // 别忘记关闭
    ctx := opentracing.ContextWithSpan(context.Background(), span) // 将span存入ctx
    session := c.GetXorm().NewSession().Context(ctx) // 设置ctx

    var a, b interface{}
    session.Sql(`select 1;`).Find(&a)
    session.Sql(`select 1;`).Find(&b)
}
```