package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDoWithoutFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := NewClient("test")
	rsp, err := c.Get(WithoutZAppFilter(context.Background()), srv.URL)
	if err != nil {
		t.Fatalf("GET 失败: %v", err)
	}
	if rsp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", rsp.StatusCode)
	}
	if rsp.Body != "ok" {
		t.Errorf("Body = %q, want ok", rsp.Body)
	}
}

func TestClientsUseIndependentRuntime(t *testing.T) {
	a, ok := NewClient("a").(*cli)
	if !ok {
		t.Fatal("NewClient 应返回 *cli")
	}
	b, ok := NewClient("b").(*cli)
	if !ok {
		t.Fatal("NewClient 应返回 *cli")
	}
	if a.std == b.std {
		t.Error("不同客户端不应共享底层运行时（连接池/代理缓存）")
	}
	if a.std.transport == a.std.insecure {
		t.Error("普通与跳过校验的 transport 不应是同一实例")
	}
}

func TestReplaceStd(t *testing.T) {
	oldClient, oldTransport := http.DefaultClient, http.DefaultTransport
	defer func() {
		http.DefaultClient, http.DefaultTransport = oldClient, oldTransport
	}()

	ReplaceStd()

	if _, ok := http.DefaultTransport.(*Transport); !ok {
		t.Fatalf("DefaultTransport 类型 = %T, want *Transport", http.DefaultTransport)
	}
	if http.DefaultClient.Transport != http.DefaultTransport {
		t.Error("DefaultClient 应复用 DefaultTransport")
	}
}

func TestWithoutZAppFilter(t *testing.T) {
	ctx := context.Background()
	if isWithoutZAppFilter(ctx) {
		t.Fatal("默认不应标记为跳过 filter")
	}
	if !isWithoutZAppFilter(WithoutZAppFilter(ctx)) {
		t.Fatal("WithoutZAppFilter 应打上标记")
	}
}
