
简单的http客户端

示例

```go
package main

import (
	"context"

	"go.uber.org/zap"

	"github.com/zly-app/component/http"
	"github.com/zly-app/uapp"
	"github.com/zly-app/zapp/logger"
)

func main() {
	app := uapp.NewApp("zapp.test")
	defer app.Exit()

	c := http.NewClient("sogou")
	rsp, err := c.Head(context.Background(), "https://sogou.com/")
	if err != nil {
		logger.Log.Fatal("请求失败", zap.Error(err))
	}
	logger.Log.Info(rsp.Body)
}
```

# 替换 http 包的的 DefaultClient 和 DefaultTransport 以默认支持相关监测

```go
func main() {
	http.ReplaceStd()
}
```
