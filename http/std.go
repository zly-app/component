package http

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/zly-app/zapp/component/conn"
)

// stdRuntime 是一个客户端实例独占的底层运行时：拨号器、两条真实 transport
// （普通 / 跳过 TLS 校验）与代理地址缓存。
//
// 它取代了原先的包级 rawStdClient / rawStdTransport / proxyConn 全局：
// 连接池归调用方持有的客户端实例所有，不同客户端互不影响。
type stdRuntime struct {
	dialer    *net.Dialer
	transport *http.Transport
	insecure  *http.Transport
	proxyConn *conn.AnyConn[*url.URL]
}

func newStdRuntime() *stdRuntime {
	s := &stdRuntime{
		dialer: &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		proxyConn: conn.NewAnyConn[*url.URL](nil),
	}
	s.transport = s.newTransport(false)
	s.insecure = s.newTransport(true)
	return s
}

func (s *stdRuntime) newTransport(insecureSkipVerify bool) *http.Transport {
	t := &http.Transport{
		Proxy:                 s.proxyResolve,
		DialContext:           s.dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if insecureSkipVerify {
		t.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // 跳过tls校验
			RootCAs:            x509.NewCertPool(),
		}
	}
	return t
}

func (s *stdRuntime) proxyResolve(req *http.Request) (*url.URL, error) {
	proxy := getProxyByCtx(req.Context())
	if proxy == "" {
		return http.ProxyFromEnvironment(req)
	}
	return s.proxyConn.GetConn(url.Parse, proxy)
}
