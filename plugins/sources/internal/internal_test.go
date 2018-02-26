package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/internal/testutil"
	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestInternal_impl(t *testing.T) {
	var _ optic.Source = new(Internal)
}

func TestInternalPlugin(t *testing.T) {
	i := NewInternal()
	acc := &testutil.Accumulator{}

	i.Gather(acc)
	assert.True(t, acc.HasMetric("internal_memstats"))

	// test that a registered counter is incremented
	counter := selfmetric.GetOrRegisterCounter("mytest", "test", map[string]string{"test": "foo"})
	counter.Inc(1)
	counter.Inc(2)
	i.Gather(acc)
	acc.AssertContainsMetricWithTaggedFields(t, "internal_mytest",
		map[string]string{
			"test": "foo",
		},
		map[string]interface{}{
			"test": int64(3),
		},
	)
	acc.ClearEvents()

	// test that a registered counter is set properly
	counter.Inc(98)
	i.Gather(acc)
	acc.AssertContainsMetricWithTaggedFields(t, "internal_mytest",
		map[string]string{
			"test": "foo",
		},
		map[string]interface{}{
			"test": int64(101),
		},
	)
	acc.ClearEvents()

	// test that counter and gauge can share the same namespace, and that timings
	// are set properly
	histogram := selfmetric.GetOrRegisterHistogram("mytest", "test_nanoseconds", map[string]string{"test": "foo"})
	histogram.Update(100)
	histogram.Update(200)
	i.Gather(acc)
	acc.AssertContainsMetricWithTaggedFields(t, "internal_mytest",
		map[string]string{
			"test": "foo",
		},
		map[string]interface{}{
			"test":             int64(101),
			"test_nanoseconds": float64(150),
		},
	)
	acc.ClearEvents()
}
