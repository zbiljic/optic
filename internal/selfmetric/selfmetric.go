package selfmetric

import (
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zbiljic/pkg/metrics"

	"github.com/zbiljic/optic/internal"
	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/optic/metric"
)

// Namespace defines the common namespace to be used by all internal metrics.
const internalNamespace = "internal"

const (
	metricsNameSeparator = ","
)

var (
	registry metrics.Registry
	mu       sync.Mutex
)

// GetOrRegisterCounter returns an existing `metrics.Counter` or
// constructs and registers a new `metrics.Counter`.
func GetOrRegisterCounter(namespace, field string, tags map[string]string) metrics.Counter {
	mu.Lock()
	defer mu.Unlock()

	name := internal.BuildFQName(internalNamespace, namespace)
	key := key(name, tags)
	mm := metrics.GetOrRegisterMultiMetric(key, tags, registry)
	if metric, ok := mm.Metrics()[field]; ok {
		switch v := metric.(type) {
		case metrics.Counter:
			return v
		default:
			panic("Attempted to register Counter over other metric")
		}
	}
	c := mm.GetOrAdd(field, metrics.NewCounter).(metrics.Counter)
	return c
}

// GetOrRegisterGauge returns an existing `metrics.Gauge` or constructs
// and registers a new `metrics.Gauge`.
func GetOrRegisterGauge(namespace, field string, tags map[string]string) metrics.Gauge {
	mu.Lock()
	defer mu.Unlock()

	name := internal.BuildFQName(internalNamespace, namespace)
	key := key(name, tags)
	mm := metrics.GetOrRegisterMultiMetric(key, tags, registry)
	if metric, ok := mm.Metrics()[field]; ok {
		switch v := metric.(type) {
		case metrics.Gauge:
			return v
		default:
			panic("Attempted to register Gauge over other metric")
		}
	}
	c := mm.GetOrAdd(field, metrics.NewGauge).(metrics.Gauge)
	return c
}

// GetOrRegisterGaugeFloat64 returns an existing `metrics.GaugeFloat64` or
// constructs and registers a new `metrics.GaugeFloat64`.
func GetOrRegisterGaugeFloat64(namespace, field string, tags map[string]string) metrics.GaugeFloat64 {
	mu.Lock()
	defer mu.Unlock()

	name := internal.BuildFQName(internalNamespace, namespace)
	key := key(name, tags)
	mm := metrics.GetOrRegisterMultiMetric(key, tags, registry)
	if metric, ok := mm.Metrics()[field]; ok {
		switch v := metric.(type) {
		case metrics.GaugeFloat64:
			return v
		default:
			panic("Attempted to register GaugeFloat64 over other metric")
		}
	}
	c := mm.GetOrAdd(field, metrics.NewGaugeFloat64).(metrics.GaugeFloat64)
	return c
}

// GetOrRegisterHistogram returns an existing `metrics.Histogram` or
// constructs and registers a new `metrics.Histogram`.
func GetOrRegisterHistogram(namespace, field string, tags map[string]string) metrics.Histogram {
	mu.Lock()
	defer mu.Unlock()

	name := internal.BuildFQName(internalNamespace, namespace)
	key := key(name, tags)
	mm := metrics.GetOrRegisterMultiMetric(key, tags, registry)
	if metric, ok := mm.Metrics()[field]; ok {
		switch v := metric.(type) {
		case metrics.Histogram:
			return v
		default:
			panic("Attempted to register Histogram over other metric")
		}
	}
	s := metrics.NewUniformSample(10000)
	c := mm.GetOrAdd(field, func() metrics.Histogram { return metrics.NewHistogram(s) }).(metrics.Histogram)
	return c
}

func key(namespace string, tags map[string]string) string {
	k := namespace
	k += metricsNameSeparator

	tmp := make([]string, len(tags))
	i := 0
	for k, v := range tags {
		tmp[i] = k + v
		i++
	}
	sort.Strings(tmp)

	for _, s := range tmp {
		k += s
	}

	return k
}

// Metrics returns all registered stats as optic metrics.
func Metrics() []optic.Metric {
	mu.Lock()
	defer mu.Unlock()

	result := make([]optic.Metric, 0)
	registry.Each(func(name string, m metrics.Metric) {
		if idx := strings.Index(name, metricsNameSeparator); idx != -1 {
			name = name[:idx]
		}
		switch v := m.(type) {
		case metrics.MultiMetric:
			mmSnapshot := v.Snapshot()
			var (
				mtags   map[string]string
				mfields = make(map[string]interface{})
				mtime   = time.Now()
			)
			mtags = mmSnapshot.Tags()

			mmMetrics := mmSnapshot.Metrics()

			for fieldName, metricValue := range mmMetrics {
				switch mv := metricValue.(type) {
				case metrics.Counter:
					mfields[fieldName] = mv.Count()
				case metrics.Gauge:
					mfields[fieldName] = mv.Value()
				case metrics.GaugeFloat64:
					mfields[fieldName] = mv.Value()
				case metrics.Histogram:
					// NOTE: using mean value from histogram
					mfields[fieldName] = mv.Mean()
				}
			}

			metric, err := metric.NewParsed(name, mtags, mfields, mtime)
			if err != nil {
				log.Printf("ERROR Error creating selfstat metric: %s", err)
				return
			}
			result = append(result, metric)
		default:
			log.Printf("ERROR Found illegal internal metric type: %v", v)
		}
	})

	return result
}

func init() {
	registry = metrics.NewRegistry()
}
