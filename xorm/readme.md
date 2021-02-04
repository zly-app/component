
# xorm组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

```text
xorm.IXormComponent         组件接口
xorm.NewXorm                创建xorm组件

c.GetXorm(name ...string)           获取xorm引擎
``` 

# 配置

> 默认组件类型为 `xorm`

```toml
[components.xorm.default]
# 驱动
Driver="sqllite3"
# 连接源
Source="test.db"
# 最大空闲连接数
MaxIdleConns=3
# 最大连接池个数
MaxOpenConns=10
# 最大续航时间(毫秒, 0表示无限
ConnMaxLifetime=0
```
