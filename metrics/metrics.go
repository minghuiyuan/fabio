package metrics

import (
	"fmt"
	"github.com/fabiolb/fabio/config"
	gkm "github.com/go-kit/kit/metrics"
	"log"
	"strings"
)

// Provider is an abstraction of a metrics backend.
type Provider interface {
	// NewCounter creates a new counter object.
	NewCounter(name string, labels ...string) gkm.Counter

	// NewGauge creates a new gauge object.
	NewGauge(name string, labels ...string) gkm.Gauge

	// NewHistogram creates a new histogram object
	NewHistogram(name string, labels ...string) gkm.Histogram

	// Unregister removes a previously registered
	// name or metric. Required for go-metrics and
	// service pruning. This signature is probably not
	// correct.
	Unregister(v interface{})
}

func Initialize(cfg *config.Metrics) (Provider, error) {
	var p []Provider
	var prefix string
	var err error
	if prefix, err = parsePrefix(cfg.Prefix); err != nil {
		return nil, fmt.Errorf("metrics: invalid Prefix template: %w", err)
	}
	for _, x := range strings.Split(cfg.Target, ",") {
		x = strings.TrimSpace(x)
		switch x {
		case "flat":
			p = append(p, &flatProvider{prefix})
		case "label":
			p = append(p, &labelProvider{prefix})
		case "statsd":
			pp, err := newStatsdProvider(prefix, cfg.StatsDAddr, cfg.Interval)
			if err != nil {
				return nil, err
			}
			p = append(p, pp)
		case "prometheus":
			p = append(p, newPromProvider(prefix, cfg.Prometheus.Subsystem, cfg.Prometheus.Buckets))

		default:
			log.Printf("[WARN] Skipping unknown metrics provider %q", x)
			continue
		}
		log.Printf("[INFO] Registering metrics provider %q", x)

		if len(p) == 0 {
			log.Printf("[INFO] Metrics disabled")
		}
	}
	if len(p) == 0 {
		return &DiscardProvider{}, nil
	}
	return NewMultiProvider(p), nil
}
