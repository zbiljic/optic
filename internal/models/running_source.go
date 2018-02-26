package models

import (
	"fmt"
	"log"
	"time"

	"github.com/zbiljic/pkg/metrics"

	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/optic/logline"
	"github.com/zbiljic/optic/optic/metric"
	"github.com/zbiljic/optic/optic/raw"
)

const (
	defaultEventChannelBufferSize = 100
)

var (
	GlobalEventsProcessed = selfmetric.GetOrRegisterCounter("agent", "events_processed", map[string]string{})
)

type RunningSource struct {
	Source optic.Source
	Config *SourceConfig

	trace       bool // only used by 'test' command
	defaultTags map[string]string

	eventsCh chan optic.Event

	forwardFunc func(optic.Event)

	EventsProcessed metrics.Counter
}

func NewRunningSource(source optic.Source, config *SourceConfig) *RunningSource {
	r := &RunningSource{
		Source:   source,
		Config:   config,
		eventsCh: make(chan optic.Event, defaultEventChannelBufferSize),
		EventsProcessed: selfmetric.GetOrRegisterCounter(
			"sources",
			"events_processed",
			map[string]string{"source": config.Name},
		),
	}

	r.forwardFunc = forwardFunc(
		r.Name(),
		r.Config.ForwardProcessors,
		r.Config.ForwardSinks,
	)

	if r.Config.Decoder != nil {
		// configure decoder if possible
		if di, ok := r.Source.(optic.DecoderInput); ok {
			di.SetDecoder(r.Config.Decoder)
		}
	}

	return r
}

// SourceConfig containing a kind, name, and interval.
type SourceConfig struct {
	Kind string
	Name string

	Interval time.Duration
	Tags     map[string]string

	Decoder optic.Decoder

	Processors []*RunningProcessor

	ForwardProcessors []*RunningProcessor
	ForwardSinks      []*RunningSink
}

func (r *RunningSource) Name() string {
	return "sources." + r.Config.Name
}

func (r *RunningSource) Trace() bool {
	return r.trace
}

func (r *RunningSource) SetTrace(trace bool) {
	r.trace = trace
}

func (r *RunningSource) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}

func (r *RunningSource) EventsCh() chan optic.Event {
	return r.eventsCh
}

// ForwardEvent adds an event to the source to be forwarded.
func (r *RunningSource) ForwardEvent(event optic.Event) {
	if event == nil {
		return
	}

	r.forwardFunc(event)
}

func (r *RunningSource) MakeRaw(
	source string,
	value []byte,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) optic.Raw {

	raw := makeRaw(
		source,
		value,
		tags,
		fields,
		t,
		r.Config.Tags,
		r.defaultTags,
	)

	if r.trace && raw != nil {
		fmt.Print("> " + raw.String())
	}

	GlobalEventsProcessed.Inc(1)
	r.EventsProcessed.Inc(1)

	return raw
}

func (r *RunningSource) MakeMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	metricType optic.MetricType,
	t time.Time,
) optic.Metric {

	metric := makeMetric(
		name,
		tags,
		fields,
		metricType,
		t,
		r.Config.Tags,
		r.defaultTags,
	)

	if r.trace && metric != nil {
		fmt.Print("> " + metric.String())
	}

	GlobalEventsProcessed.Inc(1)
	r.EventsProcessed.Inc(1)

	return metric
}

func (r *RunningSource) MakeLogLine(
	path string,
	content string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
) optic.LogLine {

	logline := makeLogLine(
		path,
		content,
		tags,
		fields,
		t,
		r.Config.Tags,
		r.defaultTags,
	)

	if r.trace && logline != nil {
		fmt.Print("> " + logline.String())
	}

	GlobalEventsProcessed.Inc(1)
	r.EventsProcessed.Inc(1)

	return logline
}

func makeRaw(
	source string,
	value []byte,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
	pluginTags map[string]string,
	daemonTags map[string]string,
) optic.Raw {

	if tags == nil {
		tags = make(map[string]string)
	}
	if fields == nil {
		fields = make(map[string]interface{})
	}

	// Apply plugin-wide tags if set
	for k, v := range pluginTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	// Apply daemon-wide tags if set
	for k, v := range daemonTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	r, err := raw.New(source, value, tags, fields, t)
	if err != nil {
		log.Printf("WARN Error adding raw [%s]: %s\n", source, err.Error())
		return nil
	}

	return r
}

func makeMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	metricType optic.MetricType,
	t time.Time,
	pluginTags map[string]string,
	daemonTags map[string]string,
) optic.Metric {
	if len(fields) == 0 || len(name) == 0 {
		return nil
	}
	if tags == nil {
		tags = make(map[string]string)
	}

	// Apply plugin-wide tags if set
	for k, v := range pluginTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	// Apply daemon-wide tags if set
	for k, v := range daemonTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	for k, v := range fields {
		if !metric.MetricNameValid(k) {
			log.Printf("DEBUG Metric [%s] field name [%s] invalid, skipping", name, k)
			delete(fields, k)
			continue
		}
		// Validate uint64 and float64 fields
		// convert all int & uint types to int64
		switch val := v.(type) {
		case nil:
			// delete nil fields
			delete(fields, k)
		case uint:
			fields[k] = int64(val)
			continue
		case uint8:
			fields[k] = int64(val)
			continue
		case uint16:
			fields[k] = int64(val)
			continue
		case uint32:
			fields[k] = int64(val)
			continue
		case int:
			fields[k] = int64(val)
			continue
		case int8:
			fields[k] = int64(val)
			continue
		case int16:
			fields[k] = int64(val)
			continue
		case int32:
			fields[k] = int64(val)
			continue
		case uint64:
			if val < uint64(9223372036854775808) {
				fields[k] = int64(val)
			} else {
				fields[k] = int64(9223372036854775807)
			}
			continue
		case float32:
			fields[k] = float64(val)
			continue
		case float64:
			fields[k] = v
		case string:
			fields[k] = v
		default:
			fields[k] = v
		}
	}

	m, err := metric.New(name, tags, fields, t, metricType)
	if err != nil {
		log.Printf("WARN Error adding metric [%s]: %s\n", name, err.Error())
		return nil
	}

	return m
}

func makeLogLine(
	path string,
	content string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
	pluginTags map[string]string,
	daemonTags map[string]string,
) optic.LogLine {

	if tags == nil {
		tags = make(map[string]string)
	}
	if fields == nil {
		fields = make(map[string]interface{})
	}

	// Apply plugin-wide tags if set
	for k, v := range pluginTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	// Apply daemon-wide tags if set
	for k, v := range daemonTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}

	ll, err := logline.New(path, content, tags, fields, t)
	if err != nil {
		log.Printf("WARN Error adding logline [%s]: %s\n", content, err.Error())
		return nil
	}

	return ll
}
