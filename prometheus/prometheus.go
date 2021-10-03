package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zly-app/zapp/core"
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
	return p
}

func (p *Prometheus) RegistryPrometheusCounter(name, help string, constLabels Labels, labels ...string) {
	if _, ok := p.counterCollector[name]; ok {
		p.app.Fatal("重复注册prometheus计数器")
	}

	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   p.app.Name(),
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := prometheus.Register(counter)
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
		Namespace:   p.app.Name(),
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := prometheus.Register(gauge)
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
		Namespace:   p.app.Name(),
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := prometheus.Register(histogram)
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
		Namespace:   p.app.Name(),
		Subsystem:   "",
		Name:        name,
		Help:        help,
		ConstLabels: constLabels,
	}, labels)
	err := prometheus.Register(summary)
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
