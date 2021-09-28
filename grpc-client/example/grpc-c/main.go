package main

import (
	"context"
	"fmt"

	grpc_client "github.com/zly-app/component/grpc-client"
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"

	"grpc-c/pb/hello"
)

type Component struct {
	core.IComponent
	grpc_client.IGrpcClient
}

func (c *Component) Close() {
	c.IComponent.Close()
	c.IGrpcClient.Close()
}

func main() {
	app := zapp.NewApp("grpc-c",
		zapp.WithCustomComponent(func(app core.IApp) core.IComponent { // 自定义组件
			return &Component{
				IComponent:  app.GetComponent(),
				IGrpcClient: grpc_client.NewGrpcClient(app), // 创建grpc客户端
			}
		}))
	defer app.Exit()

	c := app.GetComponent().(*Component)
	// 注册客户端创造者
	c.RegistryGrpcClientCreator("hello", hello.NewHelloServiceClient)
	// 获取客户端
	httpClient := c.GetGrpcClient("hello").(hello.HelloServiceClient)

	// 调用
	resp, err := httpClient.Hello(context.Background(), &hello.HelloReq{Msg: "hello"})
	if err != nil {
		app.Fatal(resp)
	}
	fmt.Println(resp.GetMsg())
}
