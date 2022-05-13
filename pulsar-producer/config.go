package pulsar_producer

import (
	"github.com/zly-app/zapp/core"
)

const DefaultComponentType core.ComponentType = "pulsar-producer"

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (conf *Config) Check() error {
	return nil
}
