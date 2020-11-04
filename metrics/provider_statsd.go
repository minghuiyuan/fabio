package metrics

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/statsd"
)

type statsdProvider struct {
	s  *statsd.Statsd
	t  *time.Ticker
}

func newStatsdProvider(prefix, addr string, interval time.Duration) (*statsdProvider, error) {
	p := &statsdProvider{
		s: statsd.New(prefix, log.NewNopLogger()),
	}
	_, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving statsd address %s: %w", addr, err)
	}
	p.t = time.NewTicker(interval)
	go func() {
		p.s.SendLoop(context.Background(), p.t.C, "udp", addr)
	}()

	return p, nil
}

type statsdCounter struct {
	gkm.Counter
	routeCounter bool
	name         string
	p            *statsdProvider
	labels       []string
}

func (c *statsdCounter) With(labelValues ...string) gkm.Counter {
	var name string

	switch c.routeCounter {
	case true:
		var err error
		name, err = TargetNameWith(c.name, c.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(c.name, labelValues, DotSeparator)
	}
	return &statsdCounter{
		Counter:      c.p.s.NewCounter(name, 1),
		name:         name,
		labels:       c.labels,
		routeCounter: c.routeCounter,
	}
}

// NewCounter - This assumes if there are labels, there will be a With() call
func (p *statsdProvider) NewCounter(name string, labels ...string) gkm.Counter {
	if len(labels) == 0 {
		return p.s.NewCounter(name, 1)
	}
	rc := strings.HasPrefix(name, RoutePrefix)
	return &statsdCounter{
		name:         name,
		p:            p,
		labels:       labels,
		routeCounter: rc,
	}

}

type statsdGauge struct {
	gkm.Gauge
	name       string
	p          *statsdProvider
	labels     []string
	routeGauge bool
}

func (g *statsdGauge) With(labelValues ...string) gkm.Gauge {
	var name string
	switch g.routeGauge {
	case true:
		var err error
		name, err = TargetNameWith(g.name, g.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(g.name, labelValues, DotSeparator)
	}
	return &statsdGauge{
		Gauge:      g.p.s.NewGauge(name),
		name:       name,
		p:          g.p,
		labels:     g.labels,
		routeGauge: g.routeGauge,
	}
}

// NewGauge - this assumes if there are labels, there will be a With() call.
func (p *statsdProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	if len(labels) == 0 {
		return p.s.NewGauge(name)
	}
	rc := strings.HasPrefix(name, RoutePrefix)
	return &statsdGauge{
		name:       name,
		labels:     labels,
		routeGauge: rc,
	}
}

type statsdHistogram struct {
	gkm.Histogram
	p              *statsdProvider
	name           string
	labels         []string
	routeHistogram bool
}

func (h *statsdHistogram) With(labelValues ...string) gkm.Histogram {
	var name string
	switch h.routeHistogram {
	case true:
		var err error
		name, err = TargetNameWith(h.name, h.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(h.name, labelValues, DotSeparator)
	}
	return &statsdHistogram{
		Histogram:      h.p.s.NewTiming(name, 1),
		name:           name,
		labels:         h.labels,
		routeHistogram: h.routeHistogram,
	}
}

func (h *statsdHistogram) Observe(value float64) {
	h.Histogram.Observe(value * 1000.0)
}

// NewHistogram - this assumes if there are labels, there will be a With() call.
func (p *statsdProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	if len(labels) == 0 {
		return p.s.NewTiming(name, 1)
	}
	rc := strings.HasPrefix(name, RoutePrefix)
	return &statsdHistogram{
		name:           name,
		labels:         labels,
		routeHistogram: rc,
	}
}

