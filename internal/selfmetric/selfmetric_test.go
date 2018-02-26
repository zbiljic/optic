package selfmetric

import (
	"strings"
	"sync"
	"testing"

	"github.com/zbiljic/pkg/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/zbiljic/optic/internal/testutil"
)

var (
	// only allow one test at a time
	// this is because we are dealing with a global registry
	testLock sync.Mutex
)

// testCleanup resets the global registry for test cleanup & unlocks the test lock
func testCleanup() {
	registry = metrics.NewRegistry()
	testLock.Unlock()
}

func TestRegisterCounter(t *testing.T) {
	testLock.Lock()
	defer testCleanup()

	sm1 := GetOrRegisterCounter("test", "test_field1", map[string]string{"test": "foo"})
	c2 := GetOrRegisterCounter("test", "test_field2", map[string]string{"test": "foo"})
	assert.Equal(t, int64(0), sm1.Count())

	sm1.Inc(10)
	sm1.Inc(5)
	assert.Equal(t, int64(15), sm1.Count())

	sm1.Dec(2)
	assert.Equal(t, int64(13), sm1.Count())

	c2.Inc(101)
	assert.Equal(t, int64(101), c2.Count())

	// make sure that the same field returns the same metric
	// this one should be the same as c2.
	c2_2 := GetOrRegisterCounter("test", "test_field2", map[string]string{"test": "foo"})
	assert.Equal(t, int64(101), c2_2.Count())

	// check that tags are consistent
	registry.Each(func(name string, m metrics.Metric) {
		if idx := strings.Index(name, metricsNameSeparator); idx != -1 {
			name = name[:idx]
		}
		assert.Equal(t, "internal_test", name)
		switch v := m.(type) {
		case metrics.MultiMetric:
			assert.Equal(t, map[string]string{"test": "foo"}, v.Tags())
		}
	})
}

func TestRegisterGauge(t *testing.T) {
	testLock.Lock()
	defer testCleanup()

	g1 := GetOrRegisterGauge("test", "test_field1", map[string]string{"test": "foo"})
	g2 := GetOrRegisterGauge("test", "test_field2", map[string]string{"test": "foo"})
	assert.Equal(t, int64(0), g1.Value())

	g1.Update(12)
	assert.Equal(t, int64(12), g1.Value())

	g2.Update(101)
	assert.Equal(t, int64(101), g2.Value())

	// make sure that the same field returns the same metric
	// this one should be the same as g2.
	g2_2 := GetOrRegisterGauge("test", "test_field2", map[string]string{"test": "foo"})
	assert.Equal(t, int64(101), g2_2.Value())

	// check that tags are consistent
	registry.Each(func(name string, m metrics.Metric) {
		if idx := strings.Index(name, metricsNameSeparator); idx != -1 {
			name = name[:idx]
		}
		assert.Equal(t, "internal_test", name)
		switch v := m.(type) {
		case metrics.MultiMetric:
			assert.Equal(t, map[string]string{"test": "foo"}, v.Tags())
		}
	})
}

func TestRegisterMetricsAndVerify(t *testing.T) {
	testLock.Lock()
	defer testCleanup()

	// register two metrics with the same key
	sm1 := GetOrRegisterCounter("test_counter", "test_field1_total", map[string]string{"test": "foo"})
	sm2 := GetOrRegisterCounter("test_counter", "test_field2_total", map[string]string{"test": "foo"})
	sm1.Inc(10)
	sm2.Inc(15)
	assert.Len(t, Metrics(), 1)

	// register two more metrics with different keys
	sm3 := GetOrRegisterHistogram("test_histogram", "test_field1_nanoseconds", map[string]string{"test": "bar"})
	sm4 := GetOrRegisterHistogram("test_histogram", "test_field2_nanoseconds", map[string]string{"test": "baz"})
	sm3.Update(10)
	sm4.Update(15)
	assert.Len(t, Metrics(), 3)

	// register some more metrics
	sm5 := GetOrRegisterCounter("test", "test_field1", map[string]string{"test": "bar"})
	sm6 := GetOrRegisterCounter("test", "test_field2", map[string]string{"test": "baz"})
	GetOrRegisterCounter("test", "test_field3", map[string]string{"test": "baz"})
	sm5.Inc(10)
	sm5.Inc(18)
	sm6.Inc(15)
	assert.Len(t, Metrics(), 5)

	acc := testutil.Accumulator{}
	acc.AddMetrics(Metrics())

	// verify sm1 & sm2
	acc.AssertContainsMetricWithTaggedFields(t, "internal_test_counter",
		map[string]string{
			"test": "foo",
		},
		map[string]interface{}{
			"test_field1_total": int64(10),
			"test_field2_total": int64(15),
		},
	)

	// verify sm3
	acc.AssertContainsMetricWithTaggedFields(t, "internal_test_histogram",
		map[string]string{
			"test": "bar",
		},
		map[string]interface{}{
			"test_field1_nanoseconds": float64(10),
		},
	)

	// verify sm4
	acc.AssertContainsMetricWithTaggedFields(t, "internal_test_histogram",
		map[string]string{
			"test": "baz",
		},
		map[string]interface{}{
			"test_field2_nanoseconds": float64(15),
		},
	)

	// verify sm5
	acc.AssertContainsMetricWithTaggedFields(t, "internal_test",
		map[string]string{
			"test": "bar",
		},
		map[string]interface{}{
			"test_field1": int64(28),
		},
	)

	// verify s6 & s7
	acc.AssertContainsMetricWithTaggedFields(t, "internal_test",
		map[string]string{
			"test": "baz",
		},
		map[string]interface{}{
			"test_field2": int64(15),
			"test_field3": int64(0),
		},
	)
}

func TestKeyGeneration(t *testing.T) {

	namespace := "test"

	for want, with := range map[string]map[string]string{
		"test,fooname":            {"foo": "name"},
		"test,field1barfield2baz": {"field1": "bar", "field2": "baz"},
	} {
		if got := key(namespace, with); got != want {
			t.Errorf("got %s for key. want: %s", got, want)
		}
	}
}
