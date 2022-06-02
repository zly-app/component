
# grpc客户端组件

> 提供用于 https://github.com/zly-app/zapp 的组件

# 先决条件

1. 安装protoc编译器

   从 https://github.com/protocolbuffers/protobuf/releases 下载protoc编译器, 解压protoc.exe到$GOPATH/bin/

2. 安装 ProtoBuffer Golang 支持

   ```shell
   go install github.com/golang/protobuf/protoc-gen-go@latest`
   ```

3. 安装 ProtoBuffer GRpc Golang 支持. [文档](https://grpc.io/docs/languages/go/quickstart/)

   ```shell
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

4. 数据校验需要安装 [protoc-gen-validate](https://github.com/envoyproxy/protoc-gen-validate)

   ```shell
   go install github.com/envoyproxy/protoc-gen-validate@latest
   ```

# 快速开始

1. 创建一个项目

   ```shell
   mkdir client && cd client
   go mod init client
   ```

2. 添加 `hello/hello.proto` 文件

   ```protobuf
   syntax = 'proto3';
   package hello; // 决定proto引用路径和rpc路由
   option go_package = "client/hello/hello"; // 用于对golang包管理的定位
   
   service helloService{
      rpc Hello(HelloReq) returns (HelloResp);
   }
   
   message HelloReq{
      string msg = 1;
   }
   message HelloResp{
      string msg = 1;
   }
   ```

3. 编译 proto

   ```shell
   protoc \
   --gogoslick_out=. --gogoslick_opt=paths=source_relative \
   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
   hello/hello.proto
   ````

4. 添加 `main.go` 文件

   ```go
   package main
   
   import (
       "context"
   
       grpc_client "github.com/zly-app/component/grpc-client"
       "github.com/zly-app/zapp"
   
       "client/hello"
   )
   
   func main() {
       app := zapp.NewApp("test")
       defer app.Exit()
   
       c := grpc_client.NewGrpcClient(app)                               // 获取grpc客户端插件
       c.RegistryGrpcClientCreator("hello", hello.NewHelloServiceClient) // 注册客户端创造者
       client := c.GetGrpcClient("hello").(hello.HelloServiceClient)     // 获取客户端
   
       // 调用
       resp, err := client.Hello(context.Background(), &hello.HelloReq{Msg: "hello"})
       if err != nil {
           app.Fatal(resp)
       }
       app.Info("收到结果", resp.GetMsg())
   }
   ```

5. 运行

   ```shell
   go mod tidy && go run .
   ```

# 示例项目

+ [grpc服务端](https://github.com/zly-app/service/tree/master/grpc/example/grpc-s)
+ [grpc客户端](./example/grpc-c)

# 配置文件

添加配置文件 `configs/default.yml`. 更多配置参考[这里](./config.go)

```yaml
components:
  grpc-client:
    default:
      Address: localhost:3000
```
