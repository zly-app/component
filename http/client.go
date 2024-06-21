package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/zly-app/zapp/filter"
	"gopkg.in/yaml.v3"
)

const DefaultComponentType = "http"

type Client interface {
	Get(ctx context.Context, path string, opts ...Option) (*Response, error)
	Head(ctx context.Context, path string, opts ...Option) (*Response, error)
	Post(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error)
	Put(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error)
	Patch(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error)
	Delete(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error)
	Do(ctx context.Context, req *Request) (*Response, error)
}

type cli struct {
	Name string
}

type Request struct {
	Method string
	Path   string

	Timeout time.Duration // 超时

	InsecureSkipVerify bool // 跳过x509校验

	InHeader Header // 请求head
	InParams Values // 请求参数

	Body       string
	inJsonPtr  interface{} // 输入json
	inYamlPtr  interface{} // 输入yaml
	inStream   io.Reader
	InIsStream bool // 标记输入body是流数据, 使用者不应该主动设置这个值, 它是http库自动设置的

	outJsonPtr  interface{} // 输出json
	outYamlPtr  interface{} // 输出yaml
	OutIsStream bool        // 标记响应body是流数据
}

type Response struct {
	Body       string
	BodyStream io.ReadCloser // 注意, 读取完毕需要使用者自行调用 Close

	Status        string
	StatusCode    int
	ContentLength int64
	Header        Header
	Uncompressed  bool
}

var NewClient = func(name string) Client {
	c := cli{
		Name: name,
	}
	return c
}

var rawStdClient = &http.Client{
	Transport: rawStdTransport,
}
var rawStdClientInsecureSkipVerify = &http.Client{
	Transport: rawStdTransportInsecureSkipVerify,
}

func (c cli) Get(ctx context.Context, path string, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodGet, path, "")
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c cli) Head(ctx context.Context, path string, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodHead, path, "")
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c cli) Post(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodPost, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c cli) Put(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodPut, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c cli) Patch(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodPatch, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c cli) Delete(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodDelete, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}
func (c cli) Do(ctx context.Context, req *Request) (*Response, error) {
	return c.do(ctx, req)
}

func (c cli) do(ctx context.Context, r *Request) (*Response, error) {
	if r.Body != "" && (r.inStream != nil || r.inJsonPtr != nil || r.inYamlPtr != nil) {
		return nil, errors.New("Body, inJsonPtr, inYamlPtr and inStream are mutually exclusive")
	}
	if r.inStream != nil && (r.inJsonPtr != nil || r.inYamlPtr != nil) {
		return nil, errors.New("Body, inJsonPtr, inYamlPtr and inStream are mutually exclusive")
	}
	if r.inJsonPtr != nil && r.inYamlPtr != nil {
		return nil, errors.New("Body, inJsonPtr, inYamlPtr and inStream are mutually exclusive")
	}

	if r.OutIsStream && (r.outJsonPtr != nil || r.outYamlPtr != nil) {
		return nil, errors.New("OutIsStream, outJsonPtr and outYamlPtr are mutually exclusive")
	}
	if r.outJsonPtr != nil && r.outYamlPtr != nil {
		return nil, errors.New("OutIsStream, outJsonPtr and outYamlPtr are mutually exclusive")
	}

	r.InIsStream = r.inStream != nil
	if r.Body != "" {
		r.inStream = bytes.NewBufferString(r.Body)
	}
	if r.inJsonPtr != nil {
		body, err := sonic.ConfigStd.Marshal(r.inJsonPtr)
		if err != nil {
			return nil, err
		}
		r.Body = string(body)
		r.inStream = bytes.NewReader(body)
	}
	if r.inYamlPtr != nil {
		body, err := yaml.Marshal(r.inJsonPtr)
		if err != nil {
			return nil, err
		}
		r.Body = string(body)
		r.inStream = bytes.NewReader(body)
	}

	if isWithoutZAppFilter(ctx) {
		sp, err := c._do(ctx, r)
		if err == nil {
			err = c.unmarshal(ctx, r, sp)
		}
		return sp, err
	}

	ctx, chain := filter.GetClientFilter(ctx, DefaultComponentType, c.Name, r.Method)
	meta := filter.GetCallMeta(ctx)
	meta.AddCallersSkip(1)

	rsp, err := chain.Handle(ctx, r, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		r := req.(*Request)
		return c._do(ctx, r)
	})
	if err == nil {
		sp := rsp.(*Response)
		err = c.unmarshal(ctx, r, sp)
		return sp, err
	}
	return nil, err
}

func (c cli) _do(ctx context.Context, r *Request) (*Response, error) {
	if r.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, r.Timeout)
		defer cancel()
	}

	httpReq, err := http.NewRequestWithContext(ctx, r.Method, r.Path, r.inStream)
	if err != nil {
		return nil, err
	}
	if r.InHeader != nil {
		httpReq.Header = r.InHeader
	}
	if len(r.InParams) > 0 {
		query := httpReq.URL.Query()
		for k, v := range r.InParams {
			query[k] = append(query[k], v...)
		}
		httpReq.URL.RawQuery = query.Encode()
	}

	var httpRsp *http.Response
	if r.InsecureSkipVerify {
		httpRsp, err = rawStdClientInsecureSkipVerify.Do(httpReq)
	} else {
		httpRsp, err = rawStdClient.Do(httpReq)
	}
	if err != nil {
		return nil, err
	}

	sp := &Response{}
	sp.Status = httpRsp.Status
	sp.StatusCode = httpRsp.StatusCode
	sp.ContentLength = httpRsp.ContentLength
	sp.Header = httpRsp.Header
	sp.Uncompressed = httpRsp.Uncompressed
	if !r.OutIsStream {
		body, err := io.ReadAll(httpRsp.Body)
		defer httpRsp.Body.Close()
		if err != nil {
			return nil, err
		}
		sp.Body = string(body)
	} else {
		sp.BodyStream = httpRsp.Body
	}

	return sp, nil
}

func (c cli) unmarshal(ctx context.Context, r *Request, sp *Response) error {
	if r.OutIsStream {
		return c.unmarshalStream(ctx, r, sp)
	}

	if r.outJsonPtr != nil {
		return sonic.UnmarshalString(sp.Body, r.outJsonPtr)
	}
	if r.outYamlPtr != nil {
		return yaml.Unmarshal([]byte(sp.Body), r.outYamlPtr)
	}
	return nil
}

func (c cli) unmarshalStream(ctx context.Context, r *Request, sp *Response) error {
	if r.outJsonPtr != nil {
		return sonic.ConfigDefault.NewDecoder(sp.BodyStream).Decode(r.outJsonPtr)
	}
	if r.outYamlPtr != nil {
		return yaml.NewDecoder(sp.BodyStream).Decode(r.outYamlPtr)
	}
	return nil
}

var StdClient = newStdClient()
var StdTransport = newStdTransport()

var rawStdDialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 30 * time.Second,
}
var rawStdTransport http.RoundTripper = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DialContext:           rawStdDialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}
var rawStdTransportInsecureSkipVerify http.RoundTripper = &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DialContext:           rawStdDialer.DialContext,
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true, // 跳过tls校验
		RootCAs:            x509.NewCertPool(),
	},
}

var NewTransport = func(name string) http.RoundTripper {
	return Transport{Name: name}
}

type Transport struct {
	Name string
}
type roundTripReq struct {
	Method string
	Path   string
	Body   string

	Header Header // 请求head
	Params Values // 请求参数
	req    *http.Request
}
type roundTripResponse struct {
	Body       string
	BodyStream io.ReadCloser // 注意, 读取完毕需要使用者自行调用 Close

	Status        string
	StatusCode    int
	ContentLength int64
	Header        Header
	Uncompressed  bool
	rsp           *http.Response
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if isWithoutZAppFilter(req.Context()) {
		return rawStdTransport.RoundTrip(req)
	}

	ctx, chain := filter.GetClientFilter(req.Context(), DefaultComponentType, t.Name, req.Method)
	meta := filter.GetCallMeta(ctx)
	meta.AddCallersSkip(1)

	var body []byte
	var err error
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
	}
	r := &roundTripReq{
		Method: req.Method,
		Path:   req.URL.String(),
		Body:   string(body),
		Header: req.Header,
		Params: req.URL.Query(),
		req:    req,
	}
	rsp, err := chain.Handle(ctx, r, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		r := req.(*roundTripReq)
		httpRsp, err := rawStdTransport.RoundTrip(r.req)
		if err != nil {
			return nil, err
		}

		sp := &roundTripResponse{}
		sp.Status = httpRsp.Status
		sp.StatusCode = httpRsp.StatusCode
		sp.ContentLength = httpRsp.ContentLength
		sp.Header = httpRsp.Header
		sp.Uncompressed = httpRsp.Uncompressed

		body, err := io.ReadAll(httpRsp.Body)
		_ = httpRsp.Body.Close()
		if err != nil {
			return nil, err
		}

		httpRsp.Body = io.NopCloser(bytes.NewReader(body))

		sp.Body = string(body)
		sp.rsp = httpRsp
		return sp, nil
	})
	if err != nil {
		return nil, err
	}
	return rsp.(*roundTripResponse).rsp, nil
}

func newStdClient() *http.Client {
	return &http.Client{Transport: StdTransport}
}
func newStdTransport() http.RoundTripper {
	return Transport{Name: "std"}
}

// 替换http包的client和transport. 调用此函数前必须已执行了 zapp.NewApp
func ReplaceStd() {
	http.DefaultClient = StdClient
	http.DefaultTransport = StdTransport
}

type withoutFilterKey struct{}

// 本次调用不使用zapp的filter功能
func WithoutZAppFilter(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, withoutFilterKey{}, struct{}{})
	return ctx
}

func isWithoutZAppFilter(ctx context.Context) bool {
	return ctx.Value(withoutFilterKey{}) != nil
}
