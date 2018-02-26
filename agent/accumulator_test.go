package agent

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/optic/logline"
	"github.com/zbiljic/optic/optic/metric"
	"github.com/zbiljic/optic/optic/raw"
)

// Check the interfaces are satisfied
func TestAccumulator_impl(t *testing.T) {
	var _ optic.Accumulator = new(accumulator)
}

func TestAdd(t *testing.T) {
	now := time.Now()
	events := make(chan optic.Event, 10)
	defer close(events)
	a := NewAccumulator(&TestEventMaker{}, events)

	a.AddMetric("acctest",
		map[string]string{},
		map[string]interface{}{"value": float64(101)},
	)
	a.AddMetric("acctest",
		map[string]string{"acc": "test"},
		map[string]interface{}{"value": float64(101)},
	)
	a.AddMetric("acctest",
		map[string]string{"acc": "test"},
		map[string]interface{}{"value": float64(101)},
		now)

	testm := <-events
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-events
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-events
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", now.UnixNano()),
		actual)
}

func TestAddMetrics(t *testing.T) {
	now := time.Now()
	events := make(chan optic.Event, 10)
	defer close(events)
	a := NewAccumulator(&TestEventMaker{}, events)

	fields := map[string]interface{}{
		"usage": float64(99),
	}
	a.AddMetric("acctest",
		map[string]string{},
		fields,
	)
	a.AddMetricType("acctest",
		map[string]string{"acc": "test"},
		fields,
		optic.GaugeMetric, now)
	a.AddMetricType("acctest",
		map[string]string{"acc": "test"},
		fields,
		optic.CounterMetric, now)

	testm := <-events
	actual := testm.String()
	assert.Contains(t, actual, "acctest usage=99")

	testm = <-events
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test usage=99")

	testm = <-events
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test usage=99 %d", now.UnixNano()),
		actual)
}

func TestAccAddError(t *testing.T) {
	errBuf := bytes.NewBuffer(nil)
	log.SetOutput(errBuf)
	defer log.SetOutput(os.Stderr)

	events := make(chan optic.Event, 10)
	defer close(events)
	a := NewAccumulator(&TestEventMaker{}, events)

	a.AddError(fmt.Errorf("foo"))
	a.AddError(fmt.Errorf("bar"))
	a.AddError(fmt.Errorf("baz"))

	errs := bytes.Split(errBuf.Bytes(), []byte{'\n'})
	assert.EqualValues(t, int64(3), EventErrors.Count())
	require.Len(t, errs, 4) // 4 because of trailing newline
	assert.Contains(t, string(errs[0]), "TestPlugin")
	assert.Contains(t, string(errs[0]), "foo")
	assert.Contains(t, string(errs[1]), "TestPlugin")
	assert.Contains(t, string(errs[1]), "bar")
	assert.Contains(t, string(errs[2]), "TestPlugin")
	assert.Contains(t, string(errs[2]), "baz")
}

func TestAddGauge(t *testing.T) {
	now := time.Now()
	events := make(chan optic.Event, 10)
	defer close(events)
	a := NewAccumulator(&TestEventMaker{}, events)

	a.AddMetricType("acctest",
		map[string]string{},
		map[string]interface{}{"value": float64(101)},
		optic.GaugeMetric)
	a.AddMetricType("acctest",
		map[string]string{"acc": "test"},
		map[string]interface{}{"value": float64(101)},
		optic.GaugeMetric)
	a.AddMetricType("acctest",
		map[string]string{"acc": "test"},
		map[string]interface{}{"value": float64(101)},
		optic.GaugeMetric,
		now)

	testm := <-events
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")
	assert.Equal(t, testm.(optic.Metric).MetricType(), optic.GaugeMetric)

	testm = <-events
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")
	assert.Equal(t, testm.(optic.Metric).MetricType(), optic.GaugeMetric)

	testm = <-events
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", now.UnixNano()),
		actual)
	assert.Equal(t, testm.(optic.Metric).MetricType(), optic.GaugeMetric)
}

func TestAddCounter(t *testing.T) {
	now := time.Now()
	events := make(chan optic.Event, 10)
	defer close(events)
	a := NewAccumulator(&TestEventMaker{}, events)

	a.AddMetricType("acctest",
		map[string]string{},
		map[string]interface{}{"value": float64(101)},
		optic.CounterMetric)
	a.AddMetricType("acctest",
		map[string]string{"acc": "test"},
		map[string]interface{}{"value": float64(101)},
		optic.CounterMetric)
	a.AddMetricType("acctest",
		map[string]string{"acc": "test"},
		map[string]interface{}{"value": float64(101)},
		optic.CounterMetric,
		now)

	testm := <-events
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")
	assert.Equal(t, testm.(optic.Metric).MetricType(), optic.CounterMetric)

	testm = <-events
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")
	assert.Equal(t, testm.(optic.Metric).MetricType(), optic.CounterMetric)

	testm = <-events
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d", now.UnixNano()),
		actual)
	assert.Equal(t, testm.(optic.Metric).MetricType(), optic.CounterMetric)
}

type TestEventMaker struct {
}

func (tm *TestEventMaker) Name() string {
	return "TestPlugin"
}

func (tm *TestEventMaker) MakeRaw(
	source string,
	value []byte,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) optic.Raw {
	if r, err := raw.New(source, value, tags, fields, t); err == nil {
		return r
	}
	return nil
}

func (tm *TestEventMaker) MakeMetric(
	namespace string,
	tags map[string]string,
	fields map[string]interface{},
	metricType optic.MetricType,
	t time.Time,
) optic.Metric {
	switch metricType {
	case optic.UntypedMetric:
		if m, err := metric.New(namespace, tags, fields, t); err == nil {
			return m
		}
	case optic.CounterMetric:
		if m, err := metric.New(namespace, tags, fields, t, optic.CounterMetric); err == nil {
			return m
		}
	case optic.GaugeMetric:
		if m, err := metric.New(namespace, tags, fields, t, optic.GaugeMetric); err == nil {
			return m
		}
	}
	return nil
}

func (tm *TestEventMaker) MakeLogLine(
	path string,
	content string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) optic.LogLine {
	if ll, err := logline.New(path, content, tags, fields, t); err == nil {
		return ll
	}
	return nil
}

// Check the interfaces are satisfied
var (
	_ EventMaker = &TestEventMaker{}
)
