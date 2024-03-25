package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/zly-app/zapp/filter"
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
	Body   string

	Header          Header        // 请求head
	Params          Values        // 请求参数
	RspBodyIsStream bool          // 标记响应body是流数据
	Timeout         time.Duration // 超时
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
	c := &cli{
		Name: name,
	}
	return c
}

func (c *cli) Get(ctx context.Context, path string, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodGet, path, "")
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c *cli) Head(ctx context.Context, path string, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodHead, path, "")
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c *cli) Post(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodPost, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c *cli) Put(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodPut, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c *cli) Patch(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodPatch, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}

func (c *cli) Delete(ctx context.Context, path string, reqBody []byte, opts ...Option) (*Response, error) {
	req := NewRequest(http.MethodDelete, path, string(reqBody))
	req.applyOptions(opts...)
	return c.do(ctx, req)
}
func (c *cli) Do(ctx context.Context, req *Request) (*Response, error) {
	return c.do(ctx, req)
}

func (c *cli) do(ctx context.Context, req *Request) (*Response, error) {
	ctx, chain := filter.GetClientFilter(ctx, DefaultComponentType, c.Name, req.Method)

	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	rsp, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
		r := req.(*Request)
		var reqBody io.Reader
		if r.Body != "" {
			reqBody = bytes.NewBufferString(r.Body)
		}
		httpReq, err := http.NewRequestWithContext(ctx, r.Method, r.Path, reqBody)
		if err != nil {
			return nil, err
		}
		if r.Header != nil {
			httpReq.Header = r.Header
		}
		if len(r.Params) > 0 {
			query := httpReq.URL.Query()
			for k, v := range r.Params {
				query[k] = append(query[k], v...)
			}
			httpReq.URL.RawQuery = query.Encode()
		}

		httpRsp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return nil, err
		}

		sp := &Response{}
		sp.Status = httpRsp.Status
		sp.StatusCode = httpRsp.StatusCode
		sp.ContentLength = httpRsp.ContentLength
		sp.Header = httpRsp.Header
		sp.Uncompressed = httpRsp.Uncompressed
		if !r.RspBodyIsStream {
			body, err := io.ReadAll(httpRsp.Body)
			if err != nil {
				return nil, err
			}
			_ = httpRsp.Body.Close()
			sp.Body = string(body)
		} else {
			sp.BodyStream = httpRsp.Body
		}

		return sp, nil
	})
	if err != nil {
		return nil, err
	}
	return rsp.(*Response), nil
}
