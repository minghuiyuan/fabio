package metrics

import (
	gkm "github.com/go-kit/kit/metrics"
	prommetrics "github.com/go-kit/kit/metrics/prometheus"
	promclient "github.com/prometheus/client_golang/prometheus"
)

type promProvider struct {
	Opts    promclient.Opts
	Buckets []float64
}

func newPromProvider(namespace, subsystem string, buckets []float64) Provider {
	return &promProvider{
		Opts: promclient.Opts{
			Namespace: namespace,
			Subsystem: subsystem,
		},
		Buckets: buckets,
	}
}

func (p *promProvider) NewCounter(name string, labels ...string) gkm.Counter {
	copts := promclient.CounterOpts(p.Opts)
	copts.Name = name
	return prommetrics.NewCounterFrom(copts, labels)
}

func (p *promProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	gopts := promclient.GaugeOpts(p.Opts)
	gopts.Name = name
	return prommetrics.NewGaugeFrom(gopts, labels)
}

func (p *promProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
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

