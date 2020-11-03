package label

import (
	"fmt"
	"github.com/fabiolb/fabio/metrics/names"
	gkm "github.com/go-kit/kit/metrics"
	"math"
	"sync/atomic"
)

type Provider struct{}

func (p *Provider) NewCounter(name string, labels ...string) gkm.Counter {
	return &Counter{Name: name, Labels: labels}
}

func (p *Provider) NewGauge(name string, labels ...string) gkm.Gauge {
	return &Gauge{Name: name, Labels: labels}
}

func (p *Provider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return &Histogram{Name: name, Labels: labels}
}

func (p *Provider) Unregister(interface{}) {}

type Counter struct {
	Name   string
	Labels []string
	Values []string
	v      int64
}

func (c *Counter) With(labelValues ...string) gkm.Counter {
	cc := &Counter{
		Name:   c.Name,
		Labels: c.Labels,
		Values: make([]string, len(labelValues)),
		v:      c.v,
	}
	copy(cc.Values, labelValues)
	return cc
}

func (c *Counter) Inc() {
	v := atomic.AddInt64(&c.v, 1)
	fmt.Printf("%s:%d|c%s\n", c.Name, v, names.Labels(c.Labels, c.Values, "|#", ":", ","))
}

func (c *Counter) Add(delta float64) {
	v := atomic.AddInt64(&c.v, int64(delta))
	fmt.Printf("%s:%d|c%s\n", c.Name, v, names.Labels(c.Labels, c.Values, "|#", ":", ","))
}

type Gauge struct {
	valBits uint64
	Name    string
	Labels  []string
	Values  []string
}

func (g *Gauge) With(labelValues ...string) gkm.Gauge {
	gc := &Gauge{
		Name:   g.Name,
		Labels: g.Labels,
		Values: make([]string, len(labelValues)),
	}
	copy(gc.Values, labelValues)
	return gc
}

func (g *Gauge) Set(n float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(n))
	fmt.Printf("%s:%d|g%s\n", g.Name, int(n), names.Labels(g.Labels, g.Values, "|#", ":", ","))
}

func (g *Gauge) Add(delta float64) {
	var oldBits uint64
	var newBits uint64
	for {
		oldBits = atomic.LoadUint64(&g.valBits)
		newBits = math.Float64bits(math.Float64frombits(oldBits) + delta)
		if atomic.CompareAndSwapUint64(&g.valBits, oldBits, newBits) {
			break
		}
	}
	fmt.Printf("%s:%d|g%s\n", g.Name, int(delta), names.Labels(g.Labels, g.Values, "|#", ":", ","))
}

type Histogram struct {
	Name   string
	Labels []string
	Values []string
}

func (h *Histogram) With(labels ...string) gkm.Histogram {
	h2 := &Histogram{}
	*h2 = *h
	h2.Values = make([]string, len(labels))
	copy(h2.Values, labels)
	return h2
}

func (h *Histogram) Observe(t float64) {
	fmt.Printf("%s:%d|ms%s\n", h.Name, int64(math.Round(t*100.0)), names.Labels(h.Labels, h.Values, "|#", ":", ","))
}
