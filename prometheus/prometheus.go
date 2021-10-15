package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zlyuancn/zretry"
	"go.uber.org/zap"
)

type (
	Labels  = prometheus.Labels
	Counter = interface {
		Inc()
		Add(float64)
	}
	Gauge = interface {
		Set(float64)
		Inc()
		Dec()
		Add(float64)
		Sub(float64)
		SetToCurrentTime()
	}
	Histogram = interface {
		Observe(float64)
	}
	Summary = interface {
		Observe(float64)
	}
)

type IPrometheus interface {
	/*注册prometheus计数器, 只能在app.Run()之前使用
	  name 计数器名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryPrometheusCounter(name, help string, constLabels Labels, labels ...string)
	// 获取prometheus计数器
	GetPrometheusCounter(name string, labels Labels) Counter
	// 获取prometheus计数器
	GetPrometheusCounterWithLabelValue(name string, labelValues ...string) Counter

	/*注册prometheus计量器, 只能在app.Run()之前使用
	  name 计量器名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryPrometheusGauge(name, help string, constLabels Labels, labels ...string)
	// 获取prometheus计量器
	GetPrometheusGauge(name string, labels Labels) Gauge
	// 获取prometheus计量器
	GetPrometheusGaugeWithLabelValue(name string, labelValues ...string) Gauge

	/*注册prometheus直方图, 只能在app.Run()之前使用
	  name 直方图名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryPrometheusHistogram(name, help string, constLabels Labels, labels ...string)
	// 获取prometheus直方图
	GetPrometheusHistogram(name string, labels Labels) Histogram
	// 获取prometheus直方图
	GetPrometheusHistogramWithLabelValue(name string, labelValues ...string) Histogram

	/*注册prometheus汇总, 只能在app.Run()之前使用
	  name 直方图名, 一般为 需要检测的对象_数值类型_单位
	  help 一段描述文字
	  constLabels 固定不变的标签值, 如主机名, ip 等
	  labels 允许使用的标签, 可为nil
	*/
	RegistryPrometheusSummary(name, help string, constLabels Labels, labels ...string)
	// 获取prometheus汇总
	GetPrometheusSummary(name string, labels Labels) Summary
	// 获取prometheus汇总
	GetPrometheusSummaryWithLabelValue(name string, labelValues ...string) Summary

	// 关闭
	Close()
}

type Prometheus struct {
	app                core.IApp
	componentType      core.ComponentType
	counterCollector   map[string]*prometheus.CounterVec   // 计数器
	gaugeCollector     map[string]*prometheus.GaugeVec     // 计量器
	histogramCollector map[string]*prometheus.HistogramVec // 直方图
	summaryCollector   map[string]*prometheus.SummaryVec   // 汇总

	pullRegistry prometheus.Registerer // pull模式注册器
	pusher       *push.Pusher          // push模式推送器
}

func NewPrometheus(app core.IApp, componentType ...core.ComponentType) IPrometheus {
	p := &Prometheus{
		app:                app,
		componentType:      DefaultComponentType,
		counterCollector:   make(map[string]*prometheus.CounterVec),
		gaugeCollector:     make(map[string]*prometheus.GaugeVec),
		histogramCollector: make(map[string]*prometheus.HistogramVec),
		summaryCollector:   make(map[string]*prometheus.SummaryVec),
	}
	if len(componentType) > 0 {
		p.componentType = componentType[0]
	}

	key := fmt.Sprintf("components.%s.default", p.componentType)
	conf := newConfig()
	if app.GetConfig().GetViper().IsSet(key) {
		if err := app.GetConfig().GetViper().UnmarshalKey(key, conf); err != nil {
			app.Fatal("解析 prometheus 配置失败", zap.Error(err))
		}
	}
	conf.Check()

	p.startPullMode(conf)
	p.startPushMode(conf)

	return p
}

// 启动pull模式
func (p *Prometheus) startPullMode(conf *Config) {
	if conf.PullBind == "" {
		return
	}

	// 创建注册器
	r := prometheus.NewRegistry()
	p.pullRegistry = r
	if conf.ProcessCollector {
		r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if conf.GoCollector {
		r.MustRegister(collectors.NewGoCollector())
	}

	p.app.Info("prometheus pull模式", zap.String("bind", conf.PullBind))

	// 构建server
	handler := promhttp.InstrumentMetricHandler(r, promhttp.HandlerFor(r, promhttp.HandlerOpts{}))
	mux := http.NewServeMux()
	mux.Handle(conf.PullPath, handler)
	server := &http.Server{Addr: conf.PullBind, Handler: mux}

	// 开始监听
	go func(server *http.Server) {
		if err := server.ListenAndServe(); err != nil {
			logger.Log.Fatal("启动pull模式失败", zap.Error(err))
		}
	}(server)
}

// 启动push模式
func (p *Prometheus) startPushMode(conf *Config) {
	if conf.PushAddress == "" {
		return
	}

	// 创建推送器
	pusher := push.New(conf.PushAddress, p.app.Name())
	p.pusher = pusher
	if conf.PushInstance == "" {
		conf.PushInstance = p.app.Name()
	}
	pusher.Grouping("instance", conf.PushInstance)

	if conf.ProcessCollector {
		pusher.Collector(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}
	if conf.GoCollector {
		pusher.Collector(collectors.NewGoCollector())
	}

	// 开始推送
	go func(ctx context.Context, conf *Config, pusher *push.Pusher) {
		for {
			t := time.NewTimer(time.Duration(conf.PushTimeInterval) * time.Millisecond)
			select {
			case <-ctx.Done():
				t.Stop()
				p.push(conf, pusher) // 最后一次推送
				return
			case <-t.C:
				p.push(conf, pusher)
			}
		}
	}(p.app.BaseContext(), conf, pusher)
}

// 推送
func (p *Prometheus) push(conf *Config, pusher *push.Pusher) {
	err := zretry.DoRetry(func() error {
		return pusher.Push()
	}, time.Duration(conf.PushRetryInterval)*time.Millisecond, int32(conf.PushRetry+1), func(err error) {
		p.app.Error("prometheus状态推送失败", zap.Error(err))
	})
	if err == nil {
		p.app.Debug("prometheus状态推送成功")
	}
}

// 注册收集器
func (p *Prometheus) registryCollector(collector prometheus.Collector) error {
	if p.pullRegistry != nil {
		if err := p.pullRegistry.Register(collector); err != nil {
			return err
		}
	}
	if p.pusher != nil {
		p.pusher.Collector(collector)
	}
	return nil
}

func (p *Prometheus) RegistryPrometheusCounter(name, help string, constLabels Labels, labels ...string) {
	if _, ok := p.counterCollector[name]; ok {
		p.app.Fatal("重复注册prometheus计数器")
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(counter)
	if err != nil {
		p.app.Fatal("注册prometheus计数器失败", zap.Error(err))
	}
	p.counterCollector[name] = counter
}
func (p *Prometheus) GetPrometheusCounter(name string, labels Labels) Counter {
	coll, ok := p.counterCollector[name]
	if !ok {
		p.app.Fatal("prometheus计数器不存在", zap.String("name", name))
	}
	counter, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取prometheus计数器失败", zap.Error(err))
	}
	return counter
}
func (p *Prometheus) GetPrometheusCounterWithLabelValue(name string, labelValues ...string) Counter {
	coll, ok := p.counterCollector[name]
	if !ok {
		p.app.Fatal("prometheus计数器不存在", zap.String("name", name))
	}
	counter, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取prometheus计数器失败", zap.Error(err))
	}
	return counter
}

func (p *Prometheus) RegistryPrometheusGauge(name, help string, constLabels Labels, labels ...string) {
	if _, ok := p.gaugeCollector[name]; ok {
		p.app.Fatal("重复注册prometheus计量器")
	}

	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(gauge)
	if err != nil {
		p.app.Fatal("注册prometheus计量器失败", zap.Error(err))
	}

	p.gaugeCollector[name] = gauge

}
func (p *Prometheus) GetPrometheusGauge(name string, labels Labels) Gauge {
	coll, ok := p.gaugeCollector[name]
	if !ok {
		p.app.Fatal("prometheus计量器不存在", zap.String("name", name))
	}
	gauge, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取prometheus计量器失败", zap.Error(err))
	}
	return gauge
}
func (p *Prometheus) GetPrometheusGaugeWithLabelValue(name string, labelValues ...string) Gauge {
	coll, ok := p.gaugeCollector[name]
	if !ok {
		p.app.Fatal("prometheus计量器不存在", zap.String("name", name))
	}
	gauge, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取prometheus计量器失败", zap.Error(err))
	}
	return gauge
}

func (p *Prometheus) RegistryPrometheusHistogram(name, help string, constLabels Labels, labels ...string) {
	if _, ok := p.histogramCollector[name]; ok {
		p.app.Fatal("重复注册prometheus直方图")
	}

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(histogram)
	if err != nil {
		p.app.Fatal("注册prometheus直方图失败", zap.Error(err))
	}

	p.histogramCollector[name] = histogram
}
func (p *Prometheus) GetPrometheusHistogram(name string, labels Labels) Histogram {
	coll, ok := p.histogramCollector[name]
	if !ok {
		p.app.Fatal("prometheus直方图不存在", zap.String("name", name))
	}
	histogram, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取prometheus直方图失败", zap.Error(err))
	}
	return histogram
}
func (p *Prometheus) GetPrometheusHistogramWithLabelValue(name string, labelValues ...string) Histogram {
	coll, ok := p.histogramCollector[name]
	if !ok {
		p.app.Fatal("prometheus直方图不存在", zap.String("name", name))
	}
	histogram, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取prometheus直方图失败", zap.Error(err))
	}
	return histogram
}

func (p *Prometheus) RegistryPrometheusSummary(name, help string, constLabels Labels, labels ...string) {
	if _, ok := p.summaryCollector[name]; ok {
		p.app.Fatal("重复注册prometheus汇总")
	}

	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   "",
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := p.registryCollector(summary)
	if err != nil {
		p.app.Fatal("注册prometheus汇总失败", zap.Error(err))
	}

	p.summaryCollector[name] = summary
}
func (p *Prometheus) GetPrometheusSummary(name string, labels Labels) Summary {
	coll, ok := p.summaryCollector[name]
	if !ok {
		p.app.Fatal("prometheus汇总不存在", zap.String("name", name))
	}
	summary, err := coll.GetMetricWith(labels)
	if err != nil {
		p.app.Fatal("获取prometheus汇总失败", zap.Error(err))
	}
	return summary
}
func (p *Prometheus) GetPrometheusSummaryWithLabelValue(name string, labelValues ...string) Summary {
	coll, ok := p.summaryCollector[name]
	if !ok {
		p.app.Fatal("prometheus汇总不存在", zap.String("name", name))
	}
	summary, err := coll.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		p.app.Fatal("获取prometheus汇总失败", zap.Error(err))
	}
	return summary
}

func (p *Prometheus) Close() {}
