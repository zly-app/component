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
