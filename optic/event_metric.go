package optic

import "time"

// MetricType is an enumeration of metric types that represent a simple value.
type MetricType uint8

// Possible values for the MetricType enum.
const (
	_ MetricType = iota
	CounterMetric
	GaugeMetric
	UntypedMetric
	HistogramMetric
	SummaryMetric
)

// A Metric is a name, tagmap, timestamp, and one or more metric fields
// containing point-in-time values.
type Metric interface {
	Event

	// The name of the Metric. This is user-provided and will vary by input
	// protocol.
	Name() string
	// The specific type of the metric.
	MetricType() MetricType

	SetName(name string)
	SetPrefix(prefix string)
	SetSuffix(suffix string)

	// Split will return multiple metrics with the same timestamp for each field.
	Split() []Metric
}

type metricAccumulator interface {
	AddMetric(
		name string,
		tags map[string]string,
		fields map[string]interface{},
		ts ...time.Time)

	AddMetricType(
		name string,
		tags map[string]string,
		fields map[string]interface{},
		metricType MetricType,
		ts ...time.Time)
}
