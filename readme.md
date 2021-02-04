
# 组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 说明

> 组件的使用基本按照以下顺序

1. 定义一个类继承 `core.IComponent`, 将想要使用的组件接口也写在这个类里面
2. 这个类重写组件的 `Close()` 方法, 在这里关闭所有使用的组件
3. 在 `zapp.NewApp` 时使用 `zapp.WithCustomComponent` 选项, 提供函数返回这个类的实例指针
4. 在任何地方使用时, 将 app.GetComponent() 的结果直接类型转换为这个类的指针

# 以xorm组件示例

```go
package main

import (
	"github.com/zly-app/component/xorm"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
)

// 定义自己的组件
type Component struct {
	core.IComponent     // 保留原始组件
	xorm.IXormComponent // 继承xorm组件
	// ... 其他组件, 一般为 type.IXXXComponent
}

// 重写Close()
func (c *Component) Close() {
	c.IXormComponent.Close()
	// ... 关闭其他组件, 一般为 c.IXXXComponent.Close()
}

func main() {
	app := zapp.NewApp("test",
		zapp.WithCustomComponent(func(app core.IApp) core.IComponent { // 自定义返回自己的组件
			return &Component{
				IComponent:     app.GetComponent(), // 设置原始组件
				IXormComponent: xorm.NewXorm(app),  // 设置xorm组件
				// ... 设置其他组件, 一般为 IXXXComponent: type.NewXXX(app)
			}
		}),
	)

	c := app.GetComponent().(*Component) // 直接转换为自己的组件

	c.GetXorm() // 获取xorm组件
	// c.GetXXX() 获取其它组件, 一般为 GetXXX()
}
```
