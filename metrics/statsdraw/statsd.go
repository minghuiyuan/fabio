package statsdraw

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

	"github.com/fabiolb/fabio/metrics/names"
)

type Provider struct {
	s      *statsd.Statsd
	cancel func()
	wg     sync.WaitGroup
	t      *time.Ticker
	prefix string
}

func NewProvider(prefix, addr string, interval time.Duration) (*Provider, error) {

	p := &Provider{
		s:      statsd.New(prefix, log.NewNopLogger()),
		prefix: prefix,
	}
	_, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving statsd address %s: %w", err)
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

type Counter struct {
	gkm.Counter
	routeCounter bool
	name         string
	p            *Provider
	labels       []string
}

func (c *Counter) With(labelValues ...string) gkm.Counter {
	var name string

	switch c.routeCounter {
	case true:
		var err error
		name, err = names.RouteNameWith(c.name, c.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = names.Flatten(c.name, labelValues, names.DotSeparator)
	}
	return &Counter{
		Counter:      c.p.s.NewCounter(name, 1),
		name:         name,
		labels:       c.labels,
		routeCounter: c.routeCounter,
	}
}

func (p *Provider) NewCounter(name string, labels ...string) gkm.Counter {
	if len(labels) == 0 {
		return p.s.NewCounter(name, 1)
	}
	return &Counter{
		Counter:      p.s.NewCounter(name, 1),
		name:         name,
		p:            p,
		labels:       labels,
		routeCounter: strings.HasPrefix(name, names.RoutePrefix),
	}
}

type Gauge struct {
	gkm.Gauge
	name       string
	p          *Provider
	labels     []string
	routeGauge bool
}

func (g *Gauge) With(labelValues ...string) gkm.Gauge {
	var name string
	switch g.routeGauge {
	case true:
		var err error
		name, err = names.RouteNameWith(g.name, g.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = names.Flatten(g.name, labelValues, names.DotSeparator)
	}
	return &Gauge{
		Gauge:      g.p.s.NewGauge(name),
		name:       name,
		p:          g.p,
		labels:     g.labels,
		routeGauge: g.routeGauge,
	}
}

func (p *Provider) NewGauge(name string, labels ...string) gkm.Gauge {
	g := p.s.NewGauge(name)
	if len(labels) == 0 {
		return g
	}
	return &Gauge{
		Gauge:      g,
		name:       name,
		labels:     labels,
		routeGauge: strings.HasPrefix(name, names.RoutePrefix),
	}
}

type Histogram struct {
	gkm.Histogram
	p              *Provider
	name           string
	labels         []string
	routeHistogram bool
}

func (h *Histogram) With(labelValues ...string) gkm.Histogram {
	var name string
	switch h.routeHistogram {
	case true:
		var err error
		name, err = names.RouteNameWith(h.name, h.labels, labelValues)
		if err != nil {
			panic(err)
		}
	case false:
		name = names.Flatten(h.name, labelValues, names.DotSeparator)

	}
	return &Histogram{
		Histogram:      h.p.NewHistogram(name, h.labels...),
		name:           name,
		labels:         h.labels,
		routeHistogram: h.routeHistogram,
	}
}

func (p *Provider) NewHistogram(name string, labels ...string) gkm.Histogram {
	h := p.s.NewTiming(name, 1)
	if len(labels) == 0 {
		return h
	}
	return &Histogram{
		Histogram:      h,
		name:           name,
		labels:         labels,
		routeHistogram: strings.HasPrefix(name, names.RoutePrefix),
	}
}

func (p *Provider) Unregister(interface{}) {
	p.t.Stop()
	p.cancel()
	p.wg.Wait()
}
