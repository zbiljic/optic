package agent

import (
	"log"
	"time"

	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/optic"
)

var (
	EventErrors = selfmetric.GetOrRegisterCounter("agent", "event_errors", map[string]string{})
)

type EventMaker interface {
	Name() string

	MakeRaw(
		source string,
		value []byte,
		tags map[string]string,
		fields map[string]interface{},
		t time.Time,
	) optic.Raw

	MakeMetric(
		name string,
		tags map[string]string,
		fields map[string]interface{},
		metricType optic.MetricType,
		t time.Time,
	) optic.Metric

	MakeLogLine(
		path string,
		content string,
		tags map[string]string,
		fields map[string]interface{},
		t time.Time,
	) optic.LogLine
}

type accumulator struct {
	maker EventMaker

	events chan optic.Event

	precision time.Duration
}

func NewAccumulator(
	maker EventMaker,
	events chan optic.Event,
) optic.Accumulator {
	acc := accumulator{
		maker:     maker,
		events:    events,
		precision: time.Nanosecond,
	}
	return &acc
}

func (ac *accumulator) AddEvent(event optic.Event) {
	ac.events <- event
}

func (ac *accumulator) AddRaw(
	source string,
	value []byte,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) {
	if e := ac.maker.MakeRaw(source, value, tags, fields, ac.getTime(t)); e != nil {
		ac.events <- e
	}
}

func (ac *accumulator) AddMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) {
	if e := ac.maker.MakeMetric(name, tags, fields, optic.UntypedMetric, ac.getTime(t)); e != nil {
		ac.events <- e
	}
}

func (ac *accumulator) AddMetricType(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	metricType optic.MetricType,
	t ...time.Time,
) {
	if e := ac.maker.MakeMetric(name, tags, fields, metricType, ac.getTime(t)); e != nil {
		ac.events <- e
	}
}

func (ac *accumulator) AddLogLine(
	path string,
	content string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) {
	if e := ac.maker.MakeLogLine(path, content, tags, fields, ac.getTime(t)); e != nil {
		ac.events <- e
	}
}

// AddError passes a runtime error to the accumulator.
// The error will be tagged with the plugin name and written to the log.
func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	EventErrors.Inc(1)
	//TODO suppress/throttle consecutive duplicate errors?
	log.Printf("ERROR Error in plugin [%s]: %s", ac.maker.Name(), err)
}

func (ac accumulator) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(ac.precision)
}
