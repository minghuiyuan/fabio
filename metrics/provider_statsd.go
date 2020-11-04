package metrics

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	gkm "github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/statsd"
)

type statsdProvider struct {
	s      *statsd.Statsd
	cancel func()
	wg     sync.WaitGroup
	t      *time.Ticker
	prefix string
}

func newStatsdProvider(prefix, addr string, interval time.Duration) (*statsdProvider, error) {
	p := &statsdProvider{
		s:      statsd.New(prefix, log.NewNopLogger()),
		prefix: prefix,
	}
	_, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving statsd address %s: %w", addr, err)
	}
	var ctx context.Context
	ctx, p.cancel = context.WithCancel(context.Background())
	p.t = time.NewTicker(interval)
	p.wg.Add(1)
	go func() {
		p.s.SendLoop(ctx, p.t.C, "udp", addr)
		p.wg.Done()
	}()

	return p, nil
}

type statsdCounter struct {
	gkm.Counter
	prefix       string
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
		name, err = RouteNameWith(c.name, c.labels, labelValues)
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

func (p *statsdProvider) NewCounter(name string, labels ...string) gkm.Counter {
	if len(labels) == 0 {
		return p.s.NewCounter(name, 1)
	}
	rc := strings.HasPrefix(name, RoutePrefix)
	name = strings.Join([]string{p.prefix, name}, "--")
	return &statsdCounter{
		Counter:      p.s.NewCounter(name, 1),
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
		name, err = RouteNameWith(g.name, g.labels, labelValues)
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

func (p *statsdProvider) NewGauge(name string, labels ...string) gkm.Gauge {
	g := p.s.NewGauge(name)
	if len(labels) == 0 {
		return g
	}
	rc := strings.HasPrefix(name, RoutePrefix)
	name = strings.Join([]string{p.prefix, name}, "--")
	return &statsdGauge{
		Gauge:      g,
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
		name, err = RouteNameWith(h.name, h.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = Flatten(h.name, labelValues, DotSeparator)

	}
	return &statsdHistogram{
		Histogram:      h.p.NewHistogram(name, h.labels...),
		name:           name,
		labels:         h.labels,
		routeHistogram: h.routeHistogram,
	}
}

func (h *statsdHistogram) Observe(value float64) {
	h.Histogram.Observe(value * 1000.0)
}

func (p *statsdProvider) NewHistogram(name string, labels ...string) gkm.Histogram {
	h := p.s.NewTiming(name, 1)
	if len(labels) == 0 {
		return h
	}
	rc := strings.HasPrefix(name, RoutePrefix)
	name = strings.Join([]string{p.prefix, name}, "--")
	return &statsdHistogram{
		Histogram:      h,
		name:           name,
		labels:         labels,
		routeHistogram: rc,
	}
}

func (p *statsdProvider) Unregister(interface{}) {
	p.t.Stop()
	p.cancel()
	p.wg.Wait()
}
