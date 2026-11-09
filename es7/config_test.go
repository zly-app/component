package es7

import "testing"

func TestNewConfigDefaults(t *testing.T) {
	conf := newConfig()
	if conf.Sniff != defaultSniff {
		t.Fatalf("Sniff 默认值不符: %v", conf.Sniff)
	}
	if conf.HealthCheck != defaultHealthCheck {
		t.Fatalf("HealthCheck 默认值不符: %v", conf.HealthCheck)
	}
	if conf.Retry != defaultRetry || conf.RetryInterval != defaultRetryInterval {
		t.Fatalf("重试默认值不符: %d/%d", conf.Retry, conf.RetryInterval)
	}
	if conf.GZip != defaultGZip {
		t.Fatalf("GZip 默认值不符: %v", conf.GZip)
	}
}

func TestConfigCheckRequiresAddress(t *testing.T) {
	conf := newConfig()
	if err := conf.Check(); err == nil {
		t.Fatal("address 为空时应报错")
	}
}

func TestConfigCheckFillsDialTimeout(t *testing.T) {
	conf := newConfig()
	conf.Address = "http://localhost:9200"
	if err := conf.Check(); err != nil {
		t.Fatalf("Check 失败: %v", err)
	}
	if conf.DialTimeout != defaultDialTimeout {
		t.Fatalf("DialTimeout 未填默认值: %d", conf.DialTimeout)
	}
}

func TestConfigCheckKeepsUserValues(t *testing.T) {
	conf := newConfig()
	conf.Address = "http://a:9200,http://b:9200"
	conf.DialTimeout = 1234
	conf.Sniff = true
	conf.HealthCheck = false
	conf.Retry = 3
	conf.RetryInterval = 200
	conf.GZip = true
	if err := conf.Check(); err != nil {
		t.Fatalf("Check 失败: %v", err)
	}
	if conf.DialTimeout != 1234 || !conf.Sniff || conf.HealthCheck || conf.Retry != 3 || conf.RetryInterval != 200 || !conf.GZip {
		t.Fatalf("用户配置被覆盖: %+v", conf)
	}
}
