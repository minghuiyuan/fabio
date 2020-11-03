package flat

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/fabiolb/fabio/metrics/names"
	gkm "github.com/go-kit/kit/metrics"
)

type Provider struct{}

func (p *Provider) NewCounter(name string, labels ...string) gkm.Counter {
	return &Counter{Name: names.Flatten(name, labels, names.DotSeparator)}
}

func (p *Provider) NewGauge(name string, labels ...string) gkm.Gauge {
	return &Gauge{Name: names.Flatten(name, labels, names.DotSeparator)}
}

func (p *Provider) NewHistogram(name string, labels ...string) gkm.Histogram {
	return &Histogram{Name: names.Flatten(name, labels, names.DotSeparator)}
}

func (p *Provider) Unregister(interface{}) {}

type Counter struct {
	Name string
	v    uint64
}

func (c *Counter) With(labelValues ...string) gkm.Counter {
	return c
}

func (c *Counter) Add(v float64) {
	uv := atomic.AddUint64(&c.v, uint64(v))
	fmt.Printf("%s:%d|c\n", c.Name, uv)
}

type Gauge struct {
	// Stolen from prometheus client gauge
	valBits uint64

	Name string
}

func (g *Gauge) Set(n float64) {
	atomic.StoreUint64(&g.valBits, math.Float64bits(n))
	fmt.Printf("%s:%d|g\n", g.Name, int(n))
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
	fmt.Printf("%s:%d|g\n", g.Name, int(math.Float64frombits(newBits)))
}

func (g *Gauge) With(labelValues ...string) gkm.Gauge {
	return g
}

type Histogram struct {
	Name string
}

func (h *Histogram) Observe(t float64) {
	fmt.Printf(":%s:%d|ms\n", h.Name, int64(math.Round(t)))
}
func (h *Histogram) With(labels ...string) gkm.Histogram {
	return h
}
