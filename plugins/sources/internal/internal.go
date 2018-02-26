package internal

import (
	"runtime"

	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/plugins/sources"
)

const (
	name        = "internal"
	description = `Collect statistics about itself.`
)

type Internal struct {
	CollectMemstats bool `mapstructure:"collect_memstats"`
}

func NewInternal() optic.Source {
	return &Internal{
		CollectMemstats: true,
	}
}

func (*Internal) Kind() string {
	return name
}

func (*Internal) Description() string {
	return description
}

func (i *Internal) Gather(acc optic.Accumulator) error {
	if i.CollectMemstats {
		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)
		fields := map[string]interface{}{
			"alloc_bytes":       m.Alloc,      // bytes allocated and not yet freed
			"alloc_bytes_total": m.TotalAlloc, // bytes allocated (even if freed)
			"sys_bytes":         m.Sys,        // bytes obtained from system (sum of XxxSys below)
			"pointer_lookups":   m.Lookups,    // number of pointer lookups
			"mallocs":           m.Mallocs,    // number of mallocs
			"frees":             m.Frees,      // number of frees
			// Main allocation heap statistics.
			"heap_alloc_bytes":    m.HeapAlloc,    // bytes allocated and not yet freed (same as Alloc above)
			"heap_sys_bytes":      m.HeapSys,      // bytes obtained from system
			"heap_idle_bytes":     m.HeapIdle,     // bytes in idle spans
			"heap_in_use_bytes":   m.HeapInuse,    // bytes in non-idle span
			"heap_released_bytes": m.HeapReleased, // bytes released to the OS
			"heap_objects":        m.HeapObjects,  // total number of allocated objects
			"num_gc":              m.NumGC,
		}
		acc.AddMetric("internal_memstats", map[string]string{}, fields)
	}

	for _, m := range selfmetric.Metrics() {
		acc.AddMetric(m.Name(), m.Tags(), m.Fields(), m.Time())
	}

	return nil
}

func init() {
	sources.Add(name, NewInternal)
}
