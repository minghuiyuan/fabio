package label

import (
	"fmt"
	"github.com/fabiolb/fabio/metrics4"
	"sync/atomic"
	"time"

	"github.com/fabiolb/fabio/metrics4/names"
	gkm "github.com/go-kit/kit/metrics"

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
	v int64
}

func (c *Counter) Inc() {
	v := atomic.AddInt64(&c.v, 1)
	fmt.Printf("%s:%d|c%s\n", c.Name, v, names.Labels(c.Labels, "|#", ":", ","))
}

type Gauge struct {
	Name   string
	Labels []string
}

func (g *Gauge) Set(n float64) {
	fmt.Printf("%s:%d|g%s\n", g.Name, int(n), names.Labels(g.Labels, "|#", ":", ","))
}

type Histogram struct {
	Name string
	Labels []string
	Values []string
}

func (h *Histogram) With(labels ...string) metrics4.Histogram {
	h2 := &Histogram{}
	*h2 = *h
	h2.Values = make([]string, len(labels))
	copy(h2.Values, labels)
	return h2
}

func (h *Histogram) Observe(t time.Duration) {
	panic("implement me")
}




