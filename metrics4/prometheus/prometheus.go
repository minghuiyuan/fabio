package prometheus

import (
	gkm "github.com/go-kit/kit/metrics"
	prommetrics "github.com/go-kit/kit/metrics/prometheus"
	promclient "github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	Opts    promclient.Opts
	Buckets []float64
}

type CounterShim struct {
	*prommetrics.Counter
	realCounter *promclient.CounterVec
	labelValues []string
}

func (cs *CounterShim) With(labelValues ...string) gkm.Counter {
	cs.labelValues = make([]string, len(labelValues))
	copy(cs.labelValues, labelValues)
	return cs.Counter.With(labelValues...)
}

func (cs *CounterShim) Inc() {
	cs.realCounter.With(makeLabels(cs.labelValues...)).Inc()
}

func (p *Provider) NewCounter(name string, labels ...string) gkm.Counter {
	copts := promclient.CounterOpts(p.Opts)
	copts.Name = name
	rc := promclient.NewCounterVec(copts, labels)
	promclient.MustRegister(rc)
	return &CounterShim{
		Counter:     prommetrics.NewCounter(rc),
		realCounter: rc,
	}

}

func (p *Provider) NewGauge(name string, labels ...string) gkm.Gauge {
	gopts := promclient.GaugeOpts(p.Opts)
	gopts.Name = name
	return prommetrics.NewGaugeFrom(gopts, labels)
}

func (p *Provider) NewHistogram(name string, labels ...string) gkm.Histogram {
	hopts := promclient.HistogramOpts{
		Namespace:   p.Opts.Namespace,
		Subsystem:   p.Opts.Subsystem,
		Name:        name,
		Help:        p.Opts.Help,
		ConstLabels: p.Opts.ConstLabels,
		Buckets:     p.Buckets,
	}
	return prommetrics.NewHistogramFrom(hopts, labels)
}

func (p *Provider) Unregister(v interface{}) {
	// noop
}

func makeLabels(labelValues ...string) promclient.Labels {
	labels := promclient.Labels{}
	for i := 0; i < len(labelValues); i += 2 {
		labels[labelValues[i]] = labelValues[i+1]
	}
	return labels
}
