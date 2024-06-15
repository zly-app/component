package redis

import (
	"github.com/redis/go-redis/v9"
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

type (
	Z        = redis.Z
	ZWithKey = redis.ZWithKey
	ZStore   = redis.ZStore
)
