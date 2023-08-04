package redis

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/zly-app/zapp/pkg/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/redis/go-redis/extra/rediscmd/v9"
	"github.com/redis/go-redis/v9"
)

const (
	instrumName = "github.com/zly-app/component/redis"
)

func InstrumentTracing(rdb redis.UniversalClient) error {
	switch rdb := rdb.(type) {
	case *redis.Client:
		opt := rdb.Options()
		connString := formatDBConnString(opt.Network, opt.Addr)
		rdb.AddHook(newTracingHook(connString))
		return nil
	case *redis.ClusterClient:
		rdb.AddHook(newTracingHook(""))

		rdb.OnNewNode(func(rdb *redis.Client) {
			opt := rdb.Options()
			connString := formatDBConnString(opt.Network, opt.Addr)
			rdb.AddHook(newTracingHook(connString))
		})
		return nil
	case *redis.Ring:
		rdb.AddHook(newTracingHook(""))

		rdb.OnNewNode(func(rdb *redis.Client) {
			opt := rdb.Options()
			connString := formatDBConnString(opt.Network, opt.Addr)
			rdb.AddHook(newTracingHook(connString))
		})
		return nil
	default:
		return fmt.Errorf("redisotel: %T not supported", rdb)
	}
}

type tracingHook struct {
	tracer   trace.Tracer
	spanOpts []trace.SpanStartOption
}

var _ redis.Hook = (*tracingHook)(nil)

func newTracingHook(connString string) *tracingHook {
	t := &tracingHook{}

	t.tracer = otel.GetTracerProvider().Tracer(
		instrumName,
		trace.WithInstrumentationVersion("semver:"+redis.Version()),
	)

	t.spanOpts = []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
	}
	if connString != "" {
		t.spanOpts = append(t.spanOpts, trace.WithAttributes(semconv.DBConnectionStringKey.String(connString)))
	}
	return t
}

func (th *tracingHook) DialHook(hook redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if !trace.SpanFromContext(ctx).IsRecording() {
			return hook(ctx, network, addr)
		}

		ctx, span := th.tracer.Start(ctx, "redis.dial", th.spanOpts...)
		defer span.End()

		conn, err := hook(ctx, network, addr)
		if err != nil {
			recordError(span, err)
			return nil, err
		}
		return conn, nil
	}
}

func (th *tracingHook) ProcessHook(hook redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if !trace.SpanFromContext(ctx).IsRecording() {
			return hook(ctx, cmd)
		}

		fn, file, line := funcFileLine("github.com/redis/go-redis")

		attrs := make([]attribute.KeyValue, 0, 8)
		attrs = append(attrs,
			semconv.CodeFunctionKey.String(fn),
			semconv.CodeFilepathKey.String(file),
			semconv.CodeLineNumberKey.Int(line),
		)

		opts := th.spanOpts
		opts = append(opts, trace.WithAttributes(attrs...))

		ctx, span := th.tracer.Start(ctx, cmd.FullName(), opts...)
		defer span.End()

		cmdString := rediscmd.CmdString(cmd)
		utils.Otel.AddSpanEvent(span, "send", utils.OtelSpanKey("cmd").String(cmdString))

		if err := hook(ctx, cmd); err != nil {
			recordError(span, err)
			return err
		}

		utils.Otel.AddSpanEvent(span, "recv", utils.OtelSpanKey("val").String(getCmdVal(cmd)))
		return nil
	}
}

func (th *tracingHook) ProcessPipelineHook(
	hook redis.ProcessPipelineHook,
) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		if !trace.SpanFromContext(ctx).IsRecording() {
			return hook(ctx, cmds)
		}

		fn, file, line := funcFileLine("github.com/redis/go-redis")

		attrs := make([]attribute.KeyValue, 0, 8)
		attrs = append(attrs,
			semconv.CodeFunctionKey.String(fn),
			semconv.CodeFilepathKey.String(file),
			semconv.CodeLineNumberKey.Int(line),
			attribute.Int("db.redis.num_cmd", len(cmds)),
		)

		opts := th.spanOpts
		opts = append(opts, trace.WithAttributes(attrs...))

		ctx, span := th.tracer.Start(ctx, "redis.pipeline", opts...)
		defer span.End()

		cmdStrings := make([]string, 1, len(cmds)+1)
		for _, c := range cmds {
			cmdStrings = append(cmdStrings, rediscmd.CmdString(c))
		}
		utils.Otel.AddSpanEvent(span, "send", utils.OtelSpanKey("cmd").String(strings.Join(cmdStrings, "\n")))

		if err := hook(ctx, cmds); err != nil {
			recordError(span, err)
			return err
		}

		cmdStrings = make([]string, 1, len(cmds)+1)
		for _, c := range cmds {
			cmdStrings = append(cmdStrings, rediscmd.CmdString(c)+" = "+getCmdVal(c))
		}
		utils.Otel.AddSpanEvent(span, "recv", utils.OtelSpanKey("val").String(strings.Join(cmdStrings, "\n")))
		return nil
	}
}

func recordError(span trace.Span, err error) {
	if err != redis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

func formatDBConnString(network, addr string) string {
	if network == "tcp" {
		network = "redis"
	}
	return fmt.Sprintf("%s://%s", network, addr)
}

func funcFileLine(pkg string) (string, string, int) {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	ff := runtime.CallersFrames(pcs[:n])

	var fn, file string
	var line int
	for {
		f, ok := ff.Next()
		if !ok {
			break
		}
		fn, file, line = f.Function, f.File, f.Line
		if !strings.Contains(fn, pkg) {
			break
		}
	}

	if ind := strings.LastIndexByte(fn, '/'); ind != -1 {
		fn = fn[ind+1:]
	}

	return fn, file, line
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
