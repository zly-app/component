
# 注意, 此库已弃用, 请转到 https://github.com/zly-app/zapp/tree/master/component/metrics

---

# prometheus 指标收集器组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 此组件基于模块 [github.com/prometheus/client_golang/prometheus](https://github.com/prometheus/client_golang)

# 配置

> 默认组件类型为 `prometheus`

```toml
[components.prometheus.default]
ProcessCollector = true     # 启用进程收集器
GoCollector = true          # 启用go收集器
PullBind = ":8080"          # pull模式bind地址
PushAddress = "http://127.0.0.1:9091" # push模式 pushGateway地址
```

# 示例

```go
package main

import (
	"github.com/zly-app/component/prometheus"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
)

// 自定义组件
type Component struct {
	core.IComponent
	prometheus.IPrometheus
}

// 重写关闭方法
func (c *Component) Close() {
	c.IComponent.Close()
	c.IPrometheus.Close()
}

func main() {
	app := zapp.NewApp("myapp",
		// 设置自定义组件
		zapp.WithCustomComponent(func(app core.IApp) core.IComponent {
			return &Component{
				IComponent:  app.GetComponent(),
				IPrometheus: prometheus.NewPrometheus(app),
			}
		}),
	)

	c := app.GetComponent().(*Component) // 获取自定义组件

	// 注册计数器
	c.RegistryPrometheusCounter("myapp_logger_level_num", "测试", nil, "level")

	c.GetPrometheusCounter("myapp_logger_level_num", prometheus.Labels{"level": "info"}).Inc()
	c.GetPrometheusCounter("myapp_logger_level_num", prometheus.Labels{"level": "debug"}).Inc()
	app.Run()
}
```
