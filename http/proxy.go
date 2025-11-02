package http

import (
	"net/http"
	"net/url"

	"github.com/zly-app/zapp/component/conn"
	"golang.org/x/net/context"
)

var proxyConn = conn.NewAnyConn[*url.URL](nil)

type proxyFlag struct{}

func saveProxy2Ctx(ctx context.Context, proxy string) context.Context {
	return context.WithValue(ctx, proxyFlag{}, proxy)
}

func getProxyByCtx(ctx context.Context) string {
	value := ctx.Value(proxyFlag{})
	proxy, _ := value.(string)
	return proxy
}

func proxyResolve(req *http.Request) (*url.URL, error) {
	proxy := getProxyByCtx(req.Context())
	if proxy == "" {
		return http.ProxyFromEnvironment(req)
	}

	u, err := proxyConn.GetConn(url.Parse, proxy)
	if err != nil {
		return nil, err
	}
	return u, nil
}
