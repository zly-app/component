package http

import (
	"io"
	"net/url"
	"time"
)

type Option func(*Request)

func (r *Request) applyOptions(opts ...Option) {
	for _, opt := range opts {
		opt(r)
	}
}

func NewRequest(method, path, body string) *Request {
	o := &Request{
		Method: method,
		Path:   path,
		Body:   body,
	}
	return o
}

// 设置超时
func WithTimeout(d time.Duration) Option {
	return func(r *Request) {
		r.Timeout = d
	}
}

// 设置header
func WithInHeader(header Header) Option {
	return func(r *Request) {
		r.Header = header.Clone()
	}
}

// 设置uri参数
func WithInParams(params url.Values) Option {
	return func(r *Request) {
		r.Params = params
	}
}

// 设置请求的json数据
func WithInJson(inPtr interface{}) Option {
	return func(r *Request) {
		r.inJsonPtr = inPtr
	}
}

// 设置请求的yaml数据
func WithInYaml(inPtr interface{}) Option {
	return func(r *Request) {
		r.inYamlPtr = inPtr
	}
}

// 设置请求body流
func WithInBodyStream(body io.Reader) Option {
	return func(r *Request) {
		r.inStream = body
	}
}

// 数据解析为json, 如果标记响应body是流数据, 不需要调用 BodyStream.Close()
func WithOutJson(outPtr interface{}) Option {
	return func(r *Request) {
		r.outJsonPtr = outPtr
	}
}

// 数据解析为yaml, 如果标记响应body是流数据, 不需要调用 BodyStream.Close()
func WithOutYaml(outPtr interface{}) Option {
	return func(r *Request) {
		r.outYamlPtr = outPtr
	}
}

// 标记响应body是流数据, 读取方式从Body改为BodyStream, 且读取完毕后需要调用 BodyStream.Close()
func WithOutIsStream(isStream bool) Option {
	return func(r *Request) {
		r.OutIsStream = isStream
	}
}

// 跳过x509校验
func WithInsecureSkipVerify() Option {
	return func(r *Request) {
		r.InsecureSkipVerify = true
	}
}

// 设置代理地址. 支持 http, https, socks5, socks5h. 示例: https://127.0.0.1:1080 https://user:pwd@127.0.0.1:1080
func WithProxy(proxy string) Option {
	return func(r *Request) {
		r.Proxy = proxy
	}
}
