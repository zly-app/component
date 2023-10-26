package redis

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/extra/rediscmd/v9"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"

	"github.com/zly-app/zapp/filter"
)

func InstrumentTracing(clientType, dbname string, rdb redis.UniversalClient) error {
	switch rdb := rdb.(type) {
	case *redis.Client:
		rdb.AddHook(newTracingHook(clientType, dbname))
		return nil
	case *redis.ClusterClient:
		rdb.AddHook(newTracingHook(clientType, dbname))

		rdb.OnNewNode(func(rdb *redis.Client) {
			rdb.AddHook(newTracingHook(clientType, dbname))
		})
		return nil
	case *redis.Ring:
		rdb.AddHook(newTracingHook(clientType, dbname))

		rdb.OnNewNode(func(rdb *redis.Client) {
			rdb.AddHook(newTracingHook(clientType, dbname))
		})
		return nil
	default:
		return fmt.Errorf("redisotel: %T not supported", rdb)
	}
}

type tracingHook struct {
	clientType, clientName string
}

var _ redis.Hook = (*tracingHook)(nil)

func newTracingHook(clientType, clientName string) *tracingHook {
	t := &tracingHook{}

	t.clientType = clientType
	t.clientName = clientName
	return t
}

type dialReq struct {
	Network, Addr string
}
type dialRsp struct {
	conn net.Conn
}

func (th *tracingHook) DialHook(hook redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		ctx, chain := filter.GetClientFilter(ctx, th.clientType, th.clientName, "Dial")
		meta := filter.GetCallMeta(ctx)
		meta.AddCallersSkip(3)
		req := &dialReq{
			Network: network,
			Addr:    addr,
		}
		rsp, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			r := req.(*dialReq)
			conn, err := hook(ctx, r.Network, r.Addr)
			if err != nil {
				return nil, err
			}
			return &dialRsp{conn: conn}, nil
		})
		if err != nil {
			return nil, err
		}
		return rsp.(*dialRsp).conn, nil
	}
}

type cmdReq struct {
	cmd       redis.Cmder
	CmdString string
}
type cmdRsp struct {
	Result string
}

func (th *tracingHook) ProcessHook(hook redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, chain := filter.GetClientFilter(ctx, th.clientType, th.clientName, cmd.FullName())
		meta := filter.GetCallMeta(ctx)
		meta.AddCallersSkip(3)
		req := &cmdReq{
			cmd:       cmd,
			CmdString: rediscmd.CmdString(cmd),
		}
		_, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			r := req.(*cmdReq)
			err = hook(ctx, r.cmd)
			if err != nil {
				return nil, err
			}
			val := getCmdVal(cmd)
			sp := &cmdRsp{Result: val}
			return sp, nil
		})
		return err
	}
}

type pipeReq struct {
	cmds       []redis.Cmder
	CmdNums    int
	CmdStrings []string
}
type pipeRsp struct {
	Result []string
}

func (th *tracingHook) ProcessPipelineHook(
	hook redis.ProcessPipelineHook,
) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		ctx, chain := filter.GetClientFilter(ctx, th.clientType, th.clientName, "pipeline")
		meta := filter.GetCallMeta(ctx)
		meta.AddCallersSkip(4)
		cmdStrings := make([]string, len(cmds))
		for i, c := range cmds {
			cmdStrings[i] = rediscmd.CmdString(c)
		}
		req := &pipeReq{
			cmds:       cmds,
			CmdNums:    len(cmds),
			CmdStrings: cmdStrings,
		}
		_, err := chain.Handle(ctx, req, func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			r := req.(*pipeReq)
			err = hook(ctx, r.cmds)
			if err != nil {
				return nil, err
			}
			cmdVals := make([]string, len(cmds))
			for i, c := range cmds {
				cmdVals[i] = getCmdVal(c)
			}
			return &pipeRsp{Result: cmdVals}, nil
		})
		return err
	}
}

type ICmdValInterface interface {
	Val() interface{}
}
type ICmdValInterfaces interface {
	Val() []interface{}
}
type ICmdValString interface {
	Val() string
}
type ICmdValStrings interface {
	Val() []string
}
type ICmdValInt64 interface {
	Val() int64
}
type ICmdValInt64s interface {
	Val() []int64
}
type ICmdValDurationCmd interface {
	Val() time.Duration
}
type ICmdValTime interface {
	Val() time.Time
}
type ICmdValBool interface {
	Val() bool
}
type ICmdValBools interface {
	Val() []bool
}
type ICmdValFloat64 interface {
	Val() float64
}
type ICmdValFloat64s interface {
	Val() []float64
}
type ICmdValKVs interface {
	Val() []redis.KeyValue
}
type ICmdValMapString interface {
	Val() map[string]string
}
type ICmdValMapInt64 interface {
	Val() map[string]int64
}
type ICmdValMapStruct interface {
	Val() map[string]struct{}
}
type ICmdValZs interface {
	Val() []redis.Z
}

func getCmdVal(cmd redis.Cmder) string {
	switch v := cmd.(type) {
	case ICmdValInterface:
		return cast.ToString(v.Val())
	case ICmdValInterfaces:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValString:
		return cast.ToString(v.Val())
	case ICmdValStrings:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValInt64:
		return cast.ToString(v.Val())
	case ICmdValInt64s:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValDurationCmd:
		return cast.ToString(int64(v.Val())) + "," + v.Val().String()
	case ICmdValTime:
		return cast.ToString(v.Val().UnixNano()) + "," + cast.ToString(v.Val())
	case ICmdValBool:
		return cast.ToString(v.Val())
	case ICmdValBools:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValFloat64:
		return cast.ToString(v.Val())
	case ICmdValFloat64s:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValKVs:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v.Key))
			buff.WriteByte('=')
			buff.WriteString(cast.ToString(v.Value))
		}
		return buff.String()
	case ICmdValMapString:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for k, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(k))
			buff.WriteByte('=')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValMapInt64:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for k, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(k))
			buff.WriteByte('=')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	case ICmdValMapStruct:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for k := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(k))
		}
		return buff.String()
	case ICmdValZs:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v.Member))
			buff.WriteByte('=')
			buff.WriteString(cast.ToString(v.Score))
		}
		return buff.String()
	case *redis.Cmd:
		return cast.ToString(v.Val())
	case *redis.SliceCmd:
		vv := v.Val()
		if len(vv) == 0 {
			return ""
		}

		var buff bytes.Buffer
		for _, v := range vv {
			buff.WriteByte('\n')
			buff.WriteString(cast.ToString(v))
		}
		return buff.String()
	}
	return cmd.String()
}
