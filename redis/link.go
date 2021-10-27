package redis

import (
	"github.com/go-redis/redis/v8"
)

const (
	Nil         = redis.Nil
	KeepTTL     = redis.KeepTTL
	TxFailedErr = redis.TxFailedErr
)

var ErrClosed = redis.ErrClosed

type (
	Pipeliner       = redis.Pipeliner
	UniversalClient = redis.UniversalClient
	Cmder           = redis.Cmder
	Cmdable         = redis.Cmdable
	Error           = redis.Error
	StatefulCmdable = redis.StatefulCmdable
	ConsistentHash  = redis.ConsistentHash
	Hook            = redis.Hook
	Limiter         = redis.Limiter
)
