package main

import (
	"context"

	"github.com/zly-app/zapp"

	grpc_client "github.com/zly-app/component/grpc-client"

	"github.com/zly-app/component/grpc-client/example/grpc-c/pb/hello"
)

func main() {
	app := zapp.NewApp("grpc-c")
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
