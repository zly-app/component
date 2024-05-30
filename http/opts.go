package http

import (
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

// 设置header
func WithHeader(header Header) Option {
	return func(r *Request) {
		r.Header = header
	}
}

// 设置uri参数
func WithParams(params url.Values) Option {
	return func(r *Request) {
		r.Params = params
	}
}

// 标记响应body是流数据
func WithRspBodyIsStream(isStream bool) Option {
	return func(r *Request) {
		r.RspBodyIsStream = isStream
	}
}

// 设置超时
func WithTimeout(d time.Duration) Option {
	return func(r *Request) {
		r.Timeout = d
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

// 数据解析为json
func WithOutJson(outPtr interface{}) Option {
	return func(r *Request) {
		r.outJsonPtr = outPtr
	}
}

// 数据解析为yaml
func WithOutYaml(outPtr interface{}) Option {
	return func(r *Request) {
		r.outYamlPtr = outPtr
	}
}

// 跳过x509校验
func WithInsecureSkipVerify() Option {
	return func(r *Request) {
		r.InsecureSkipVerify = true
	}
}
