package mqtt_producer

import (
	"github.com/zly-app/zapp/core"
)

const DefaultComponentType core.ComponentType = "mqtt-producer"

const (
	defServer              = "localhost:1883"
	defWaitConnectedTimeMs = 5000
)

type Config struct {
	Server              string // mqtt服务地址. 如 localhost:1883
	WaitConnectedTimeMs int    // 等待连接超时时间, 单位毫秒

	User     string // mqtt用户名
	Password string // mqtt密码
}

func NewConfig() *Config {
	return &Config{}
}

func (conf *Config) Check() error {
	if conf.Server == "" {
		conf.Server = defServer
	}
	if conf.WaitConnectedTimeMs < 1 {
		conf.WaitConnectedTimeMs = defWaitConnectedTimeMs
	}
	return nil
}
