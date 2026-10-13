package redis

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/zly-app/zapp/log"
)

// expandHomePath 将路径前缀 ~ 展开为用户主目录（跨平台：Linux 为 $HOME，Windows 为 %USERPROFILE%）。
// 展开失败时原样返回，由后续文件读取报错。
func expandHomePath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

func NewClient(conf *RedisConfig, redisName ...string) (UniversalClient, error) {
	name := "def"
	if len(redisName) > 0 {
		name = redisName[0]
	}

	err := conf.Check()
	if err != nil {
		return nil, fmt.Errorf("redis客户端配置错误: %v", err)
	}

	clientConf := &redis.UniversalOptions{
		Addrs:                 strings.Split(conf.Address, ","),
		DB:                    conf.DB,
		Username:              conf.UserName,
		Password:              conf.Password,
		MaxRetries:            conf.MaxRetries,
		DialTimeout:           time.Duration(conf.ConnectTimeout) * time.Second,
		ReadTimeout:           time.Duration(conf.ReadTimeoutSec) * time.Second,
		WriteTimeout:          time.Duration(conf.WriteTimeoutSec) * time.Second,
		ContextTimeoutEnabled: false, // 如果设为false, 传入的ctx无效, 如果设为true, 当出现超时时会导致client关闭(因为它发送或接收了一半的命令导致client不稳定)
		PoolSize:              conf.PoolSize,
		MinIdleConns:          conf.MinIdle,
		MaxIdleConns:          conf.MaxIdle,
		ConnMaxIdleTime:       time.Duration(conf.IdleTimeout) * time.Second,
		ConnMaxLifetime:       time.Duration(conf.MaxConnLifetime) * time.Second,
		MaxRedirects:          conf.MaxRetries,
	}
	if conf.DisableClientMaintNotifications {
		clientConf.MaintNotificationsConfig = &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		}
	}
	if conf.TLSCAFile != "" {
		caCert, err := os.ReadFile(expandHomePath(conf.TLSCAFile))
		if err != nil {
			return nil, fmt.Errorf("读取CA证书失败: %v", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析CA证书失败")
		}
		tc := &tls.Config{RootCAs: caCertPool, InsecureSkipVerify: conf.InsecureSkipVerify}
		clientConf.TLSConfig = tc
	}

	var client redis.UniversalClient
	client = redis.NewUniversalClient(clientConf)

	if err := InstrumentTracing(string(DefaultComponentType), name, client); err != nil {
		log.Error("redisotel.InstrumentTracing err", err)
	}

	if err := redisotel.InstrumentMetrics(client); err != nil {
		log.Error("redisotel.InstrumentMetrics err", err)
	}
	return client, nil
}
