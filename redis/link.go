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

type (
	CommandsInfoCmd      = redis.CommandsInfoCmd
	StringSliceCmd       = redis.StringSliceCmd
	KeyFlagsCmd          = redis.KeyFlagsCmd
	StringCmd            = redis.StringCmd
	StatusCmd            = redis.StatusCmd
	IntCmd               = redis.IntCmd
	BoolCmd              = redis.BoolCmd
	DurationCmd          = redis.DurationCmd
	SliceCmd             = redis.SliceCmd
	FloatCmd             = redis.FloatCmd
	IntSliceCmd          = redis.IntSliceCmd
	ScanCmd              = redis.ScanCmd
	MapStringStringCmd   = redis.MapStringStringCmd
	KeyValueSliceCmd     = redis.KeyValueSliceCmd
	KeyValuesCmd         = redis.KeyValuesCmd
	LCSCmd               = redis.LCSCmd
	BoolSliceCmd         = redis.BoolSliceCmd
	StringStructMapCmd   = redis.StringStructMapCmd
	XMessageSliceCmd     = redis.XMessageSliceCmd
	XStreamSliceCmd      = redis.XStreamSliceCmd
	XPendingCmd          = redis.XPendingCmd
	XPendingExtCmd       = redis.XPendingExtCmd
	XAutoClaimCmd        = redis.XAutoClaimCmd
	XAutoClaimJustIDCmd  = redis.XAutoClaimJustIDCmd
	XInfoGroupsCmd       = redis.XInfoGroupsCmd
	XInfoStreamCmd       = redis.XInfoStreamCmd
	XInfoStreamFullCmd   = redis.XInfoStreamFullCmd
	XInfoConsumersCmd    = redis.XInfoConsumersCmd
	ZWithKeyCmd          = redis.ZWithKeyCmd
	ZSliceWithKeyCmd     = redis.ZSliceWithKeyCmd
	ZSliceCmd            = redis.ZSliceCmd
	FloatSliceCmd        = redis.FloatSliceCmd
	RankWithScoreCmd     = redis.RankWithScoreCmd
	ClientInfoCmd        = redis.ClientInfoCmd
	SlowLogCmd           = redis.SlowLogCmd
	TimeCmd              = redis.TimeCmd
	Cmd                  = redis.Cmd
	FunctionListCmd      = redis.FunctionListCmd
	FunctionStatsCmd     = redis.FunctionStatsCmd
	MapStringIntCmd      = redis.MapStringIntCmd
	ClusterSlotsCmd      = redis.ClusterSlotsCmd
	ClusterShardsCmd     = redis.ClusterShardsCmd
	ClusterLinksCmd      = redis.ClusterLinksCmd
	GeoPosCmd            = redis.GeoPosCmd
	GeoLocationCmd       = redis.GeoLocationCmd
	GeoSearchLocationCmd = redis.GeoSearchLocationCmd
	ACLLogCmd            = redis.ACLLogCmd
)
