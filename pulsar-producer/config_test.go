package pulsar_producer

import (
	"strings"
	"testing"
)

func TestNewConfigDefaults(t *testing.T) {
	conf := NewConfig()
	if conf.ListenerName != defListenerName {
		t.Fatalf("ListenerName 默认值不符: %q", conf.ListenerName)
	}
	if conf.DisableBlockIfQueueFull != defDisableBlockIfQueueFull || conf.DisableBatching != defDisableBatching {
		t.Fatalf("布尔默认值不符: %+v", conf)
	}
	if conf.MaxReconnectToBroker != defMaxReconnectToBroker {
		t.Fatalf("MaxReconnectToBroker 默认值不符: %d", conf.MaxReconnectToBroker)
	}
}

func TestConfigCheckFillsDefaults(t *testing.T) {
	conf := NewConfig()
	if err := conf.Check(); err != nil {
		t.Fatalf("Check 失败: %v", err)
	}
	if conf.Url != defUrl {
		t.Fatalf("Url 未填默认值: %q", conf.Url)
	}
	if conf.Topic != defTopic {
		t.Fatalf("Topic 未填默认值: %q", conf.Topic)
	}
	if conf.ConnectionTimeout != defConnectionTimeout || conf.OperationTimeout != defOperationTimeout {
		t.Fatalf("超时未填默认值: %d/%d", conf.ConnectionTimeout, conf.OperationTimeout)
	}
	if conf.SendTimeout != defSendTimeout {
		t.Fatalf("SendTimeout 未填默认值: %d", conf.SendTimeout)
	}
	if conf.MaxPendingMessages != defMaxPendingMessages {
		t.Fatalf("MaxPendingMessages 未填默认值: %d", conf.MaxPendingMessages)
	}
	if conf.CompressionType != defCompressionType || conf.CompressionLevel != defCompressionLevel {
		t.Fatalf("压缩配置未填默认值: %s/%s", conf.CompressionType, conf.CompressionLevel)
	}
	if conf.BatchingMaxPublishDelay != defBatchingMaxPublishDelay ||
		conf.BatchingMaxMessages != defBatchingMaxMessages ||
		conf.BatchingMaxSize != defBatchingMaxSize {
		t.Fatalf("批处理未填默认值: %+v", conf)
	}
	if conf.PartitionsAutoDiscoveryInterval != defPartitionsAutoDiscoveryInterval {
		t.Fatalf("分区发现间隔未填默认值: %d", conf.PartitionsAutoDiscoveryInterval)
	}
}

func TestConfigCheckSendTimeout(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"未设置取默认", 0, defSendTimeout},
		{"负数表示禁用", -5, -1},
		{"正数保留", 1234, 1234},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := NewConfig()
			conf.SendTimeout = c.in
			if err := conf.Check(); err != nil {
				t.Fatalf("Check 失败: %v", err)
			}
			if conf.SendTimeout != c.want {
				t.Fatalf("SendTimeout = %d, 期望 %d", conf.SendTimeout, c.want)
			}
		})
	}
}

func TestConfigCheckCompression(t *testing.T) {
	for _, typ := range []string{"none", "lz4", "zlib", "zstd", "LZ4"} {
		conf := NewConfig()
		conf.CompressionType = typ
		if err := conf.Check(); err != nil {
			t.Fatalf("压缩类型 %q 应被接受: %v", typ, err)
		}
	}

	conf := NewConfig()
	conf.CompressionType = "brotli"
	err := conf.Check()
	if err == nil || !strings.Contains(err.Error(), "不支持的压缩器类型") {
		t.Fatalf("非法压缩类型应报错, 实际: %v", err)
	}

	conf = NewConfig()
	conf.CompressionLevel = "fastest"
	err = conf.Check()
	if err == nil || !strings.Contains(err.Error(), "不支持的压缩等级") {
		t.Fatalf("非法压缩等级应报错, 实际: %v", err)
	}

	for _, lv := range []string{"default", "faster", "better", "BETTER"} {
		conf := NewConfig()
		conf.CompressionLevel = lv
		if err := conf.Check(); err != nil {
			t.Fatalf("压缩等级 %q 应被接受: %v", lv, err)
		}
	}
}

func TestConfigCheckKeepsUserValues(t *testing.T) {
	conf := NewConfig()
	conf.Url = "pulsar://127.0.0.1:6650"
	conf.Topic = "persistent://public/default/mine"
	conf.SendTimeout = 5000
	conf.MaxPendingMessages = 42
	conf.BatchingMaxMessages = 7
	if err := conf.Check(); err != nil {
		t.Fatalf("Check 失败: %v", err)
	}
	if conf.Url != "pulsar://127.0.0.1:6650" || conf.Topic != "persistent://public/default/mine" {
		t.Fatalf("用户地址/topic 被覆盖: %+v", conf)
	}
	if conf.SendTimeout != 5000 || conf.MaxPendingMessages != 42 || conf.BatchingMaxMessages != 7 {
		t.Fatalf("用户数值被覆盖: %+v", conf)
	}
}
