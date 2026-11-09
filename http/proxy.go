package http

import (
	"context"
)

type proxyFlag struct{}

func saveProxy2Ctx(ctx context.Context, proxy string) context.Context {
	return context.WithValue(ctx, proxyFlag{}, proxy)
}

func getProxyByCtx(ctx context.Context) string {
	value := ctx.Value(proxyFlag{})
	proxy, _ := value.(string)
	return proxy
}
