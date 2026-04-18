# HTTP 组件 - AI 参考文档

## 概述

用于 zapp 框架的 HTTP 客户端组件。提供高层客户端 API（函数式选项）和 `http.RoundTripper` 实现，自动集成 zapp filter 链。支持 OpenTelemetry 链路追踪自动传播、代理、JSON/YAML 序列化（基于 sonic）、流式请求/响应处理。

## 包路径

`github.com/zly-app/component/http`

## 核心类型

### `Client`（接口）

高层 HTTP 客户端。所有请求经过 zapp filter 链。

| 方法 | 签名 | 说明 |
|------|------|------|
| `Get` | `(ctx context.Context, path string, opts ...Option) (*Response, error)` | HTTP GET 请求 |
| `Head` | `(ctx context.Context, path string, opts ...Option) (*Response, error)` | HTTP HEAD 请求 |
| `Post` | `(ctx context.Context, path string, opts ...Option) (*Response, error)` | HTTP POST 请求 |
| `Put` | `(ctx context.Context, path string, opts ...Option) (*Response, error)` | HTTP PUT 请求 |
| `Patch` | `(ctx context.Context, path string, opts ...Option) (*Response, error)` | HTTP PATCH 请求 |
| `Delete` | `(ctx context.Context, path string, opts ...Option) (*Response, error)` | HTTP DELETE 请求 |
| `Do` | `(ctx context.Context, method string, path string, opts ...Option) (*Response, error)` | 自定义方法请求 |

### `Request`（结构体）

| 字段 | 类型 | 说明 |
|------|------|------|
| `Method` | `string` | HTTP 方法（GET/POST/PUT/PATCH/DELETE/HEAD） |
| `Path` | `string` | 请求 URL |
| `Timeout` | `time.Duration` | 请求超时时间 |
| `InsecureSkipVerify` | `bool` | 跳过 TLS x509 证书验证 |
| `Header` | `http.Header` | 请求头 |
| `Params` | `url.Values` | URL 查询参数 |
| `Body` | `string` | 请求体（字符串） |
| `inJsonPtr` | `interface{}` | 请求体 JSON 序列化（通过 sonic）— 与 Body/inYamlPtr/inStream 互斥 |
| `inYamlPtr` | `interface{}` | 请求体 YAML 序列化 — 与 Body/inJsonPtr/inStream 互斥 |
| `inStream` | `io.Reader` | 请求体流式输入 — 与 Body/inJsonPtr/inYamlPtr 互斥 |
| `outJsonPtr` | `interface{}` | 响应体 JSON 反序列化目标 — 与 outYamlPtr 互斥 |
| `outYamlPtr` | `interface{}` | 响应体 YAML 反序列化目标 — 与 outJsonPtr 互斥 |
| `OutIsStream` | `bool` | 标记响应为流式（从 BodyStream 读取而非 Body） |
| `Proxy` | `string` | 本次请求的代理 URL（http/https/socks5/socks5h） |

### `Response`（结构体）

| 字段 | 类型 | 说明 |
|------|------|------|
| `Body` | `string` | 响应体（字符串） |
| `BodyStream` | `io.ReadCloser` | 响应体流（OutIsStream=true 时可用） |
| `Status` | `string` | HTTP 状态文本 |
| `StatusCode` | `int` | HTTP 状态码 |
| `ContentLength` | `int64` | 响应内容长度 |
| `Header` | `http.Header` | 响应头 |
| `Uncompressed` | `bool` | 响应是否被解压 |

### `Transport`（结构体）

实现 `http.RoundTripper` 接口。通过 `ReplaceStd()` 使用时，所有标准库 HTTP 调用都集成 zapp filter 链。

## 选项函数

函数式选项，用于自定义请求。作为可变参数传入 Client 方法。

| 函数 | 签名 | 说明 |
|------|------|------|
| `WithTimeout` | `(d time.Duration) Option` | 设置请求超时时间 |
| `WithInHeader` | `(header http.Header) Option` | 设置请求头 |
| `WithInParams` | `(params url.Values) Option` | 设置 URL 查询参数 |
| `WithInJson` | `(inPtr interface{}) Option` | 设置请求体为 JSON（通过 sonic 序列化） |
| `WithInYaml` | `(inPtr interface{}) Option` | 设置请求体为 YAML |
| `WithInBodyStream` | `(body io.Reader) Option` | 设置请求体为流式输入 |
| `WithOutJson` | `(outPtr interface{}) Option` | 将响应体反序列化为 JSON |
| `WithOutYaml` | `(outPtr interface{}) Option` | 将响应体反序列化为 YAML |
| `WithOutIsStream` | `(isStream bool) Option` | 标记响应为流式 |
| `WithInsecureSkipVerify` | `() Option` | 跳过 TLS 证书验证 |
| `WithProxy` | `(proxy string) Option` | 设置代理 URL（http/https/socks5/socks5h） |

## 重导出类型

| 类型 | 原始类型 |
|------|----------|
| `Header` | `http.Header` |
| `Values` | `url.Values` |

## 包级函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `NewClient` | `(name string) Client` | 创建新的 HTTP 客户端实例 |
| `ReplaceStd` | `()` | 替换 `http.DefaultClient` 和 `http.DefaultTransport` 为 zapp 集成版本 |
| `WithoutZAppFilter` | `(ctx context.Context) context.Context` | 在 context 中标记跳过 zapp filter 链 |

## Filter 链集成

- **客户端 API**：请求经过 `filter.GetClientFilter()`，组件类型为 `"http"`，方法名为 HTTP 方法（GET/POST 等）
- **Transport**：通过 `ReplaceStd()` 使用时同样集成 filter 链
- **跳过 filter**：使用 `WithoutZAppFilter(ctx)` 可跳过 filter 链

## 链路追踪传播

每次请求前，客户端自动：
1. 通过 `utils.Trace.SaveToHeaders()` 将 OpenTelemetry 链路追踪上下文注入 HTTP 请求头
2. 通过 `filter.SaveCallerMeta2Header()` 附加调用方元数据（实例、环境、服务、方法）

## 代理支持

- 通过 `WithProxy("socks5://proxy:1080")` 设置单次请求代理
- 支持的协议：`http`、`https`、`socks5`、`socks5h`
- 代理 URL 通过 `conn.AnyConn` 解析并缓存复用
- 未设置代理时回退到 `http.ProxyFromEnvironment`（读取 HTTP_PROXY/HTTPS_PROXY 环境变量）

## TLS 模式

预创建两个 `http.Client` 实例：
- 标准模式：正常 TLS 证书验证
- 不安全模式：`InsecureSkipVerify=true`，通过 `WithInsecureSkipVerify()` 选项启用

## 使用模式

### 基本 GET 请求

```go
c := http.NewClient("my-service")
rsp, err := c.Get(ctx, "https://api.example.com/users")
fmt.Println(rsp.StatusCode, rsp.Body)
```

### POST JSON 请求

```go
body := map[string]interface{}{"name": "Alice", "age": 30}
var result MyResponse
rsp, err := c.Post(ctx, "https://api.example.com/users",
    http.WithInJson(body),
    http.WithOutJson(&result),
)
```

### 带超时和请求头

```go
rsp, err := c.Get(ctx, "https://api.example.com/data",
    http.WithTimeout(5*time.Second),
    http.WithInHeader(http.Header{"Authorization": {"Bearer token"}}),
    http.WithInParams(url.Values{"page": {"1"}, "size": {"10"}}),
)
```

### 流式响应

```go
rsp, err := c.Get(ctx, "https://api.example.com/stream",
    http.WithOutIsStream(true),
)
defer rsp.BodyStream.Close()
io.Copy(os.Stdout, rsp.BodyStream)
```

### 代理请求

```go
rsp, err := c.Get(ctx, "https://api.example.com",
    http.WithProxy("socks5://127.0.0.1:1080"),
)
```

### 替换标准库

```go
func main() {
    app := zapp.NewApp("test")
    defer app.Exit()
    
    http.ReplaceStd()
    // 所有标准库 http.Get/http.Post 调用现在都会经过 zapp filter 链
    resp, _ := http.Get("https://example.com")
}
```

### 跳过 Filter 链

```go
ctx = http.WithoutZAppFilter(ctx)
rsp, err := c.Get(ctx, "https://internal-service/api")
```

## 输入/输出编码规则

**输入（互斥，四选一）：**
- `Body`（string）— 原始字符串请求体
- `WithInJson` — 通过 bytedance/sonic 进行 JSON 序列化
- `WithInYaml` — YAML 序列化
- `WithInBodyStream` — 流式读取器

**输出（互斥，三选一）：**
- `WithOutJson` — 通过 sonic 进行 JSON 反序列化
- `WithOutYaml` — YAML 反序列化
- `WithOutIsStream` — 从 `Response.BodyStream` 读取

未设置任何输出选项时，`Response.Body` 包含完整的响应字符串。
