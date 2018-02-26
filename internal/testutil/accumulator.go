package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zbiljic/optic/optic"
)

// Event defines a single point measurement
type Event struct {
	Type optic.EventType

	// raw
	Source string
	Value  []byte
	// metric
	Name       string
	MetricType optic.MetricType
	// logline
	Path    string
	Content string

	Tags   map[string]string
	Fields map[string]interface{}
	Time   time.Time
}

func (e *Event) EventType() optic.EventType {
	return e.Type
}

func (e *Event) String() string {
	return fmt.Sprintf("%s %v", e.Type, e.Fields)
}

// Accumulator defines a mocked out accumulator
type Accumulator struct {
	sync.Mutex
	*sync.Cond

	Events  []*Event
	count   uint64
	Discard bool
	Errors  []error
	debug   bool
}

func (a *Accumulator) Count() uint64 {
	return atomic.LoadUint64(&a.count)
}

func (a *Accumulator) ClearEvents() {
	a.Lock()
	defer a.Unlock()
	atomic.StoreUint64(&a.count, 0)
	a.Events = make([]*Event, 0)
}

func (a *Accumulator) AddEvent(event optic.Event) {
	a.Lock()
	defer a.Unlock()
	atomic.AddUint64(&a.count, 1)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	if a.Discard {
		return
	}

	e := &Event{
		Type: event.Type(),
		Tags: event.Tags(),
		Time: event.Time(),
	}

	switch v := event.(type) {
	case optic.Raw:
		e.Source = v.Source()
		e.Value = v.Value()
		e.Fields = v.Fields()
	case optic.Metric:
		e.Name = v.Name()
		e.MetricType = v.MetricType()
		e.Fields = v.Fields()
	case optic.LogLine:
		e.Content = v.Content()
		e.Fields = v.Fields()
	}

	a.Events = append(a.Events, e)
}

func (a *Accumulator) AddEvents(events []optic.Event) {
	for _, e := range events {
		a.AddEvent(e)
	}
}

func (a *Accumulator) AddRaw(
	source string,
	value []byte,
	tags map[string]string,
	fields map[string]interface{},
	timestamp ...time.Time,
) {
	a.Lock()
	defer a.Unlock()
	atomic.AddUint64(&a.count, 1)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	if a.Discard {
		return
	}
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}

	var t time.Time
	if len(timestamp) > 0 {
		t = timestamp[0]
	} else {
		t = time.Now()
	}

	if a.debug {
		prettyTags, _ := json.MarshalIndent(tags, "", "  ")
		msg := fmt.Sprintf("Adding Raw\nTags:%s\n[%s]\n",
			string(prettyTags), string(value))
		fmt.Print(msg)
	}

	e := &Event{
		Type:   optic.RawEvent,
		Source: source,
		Value:  value,
		Tags:   tags,
		Fields: fields,
		Time:   t,
	}

	a.Events = append(a.Events, e)
}

func (a *Accumulator) AddRaws(raws []optic.Raw) {
	for _, r := range raws {
		a.AddRaw(r.Source(), r.Value(), r.Tags(), r.Fields(), r.Time())
	}
}

func (a *Accumulator) AddMetric(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	timestamp ...time.Time,
) {
	a.AddMetricType(name, tags, fields, optic.UntypedMetric, timestamp...)
}

func (a *Accumulator) AddMetricType(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	metricType optic.MetricType,
	timestamp ...time.Time,
) {
	a.Lock()
	defer a.Unlock()
	atomic.AddUint64(&a.count, 1)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	if a.Discard {
		return
	}
	if tags == nil {
		tags = map[string]string{}
	}

	if len(fields) == 0 {
		return
	}

	var t time.Time
	if len(timestamp) > 0 {
		t = timestamp[0]
	} else {
		t = time.Now()
	}

	if a.debug {
		pretty, _ := json.MarshalIndent(fields, "", "  ")
		prettyTags, _ := json.MarshalIndent(tags, "", "  ")
		msg := fmt.Sprintf("Adding Measurement [%s]\nFields:%s\nTags:%s\n",
			name, string(pretty), string(prettyTags))
		fmt.Print(msg)
	}

	e := &Event{
		Type:       optic.MetricEvent,
		Name:       name,
		MetricType: metricType,
		Tags:       tags,
		Fields:     fields,
		Time:       t,
	}

	a.Events = append(a.Events, e)
}

func (a *Accumulator) AddMetrics(metrics []optic.Metric) {
	for _, m := range metrics {
		a.AddMetric(m.Name(), m.Tags(), m.Fields(), m.Time())
	}
}

func (a *Accumulator) AddLogLine(
	path string,
	content string,
	tags map[string]string,
	fields map[string]interface{},
	timestamp ...time.Time,
) {
	a.Lock()
	defer a.Unlock()
	atomic.AddUint64(&a.count, 1)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	if a.Discard {
		return
	}
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}

	var t time.Time
	if len(timestamp) > 0 {
		t = timestamp[0]
	} else {
		t = time.Now()
	}

	if a.debug {
		prettyTags, _ := json.MarshalIndent(tags, "", "  ")
		msg := fmt.Sprintf("Adding LogLine\nTags:%s\n[%s]\n",
			string(prettyTags), content)
		fmt.Print(msg)
	}

	e := &Event{
		Type:    optic.LogLineEvent,
		Path:    path,
		Content: content,
		Tags:    tags,
		Fields:  fields,
		Time:    t,
	}

	a.Events = append(a.Events, e)
}

func (a *Accumulator) AddLogLines(loglines []optic.LogLine) {
	for _, ll := range loglines {
		a.AddLogLine(ll.Path(), ll.Content(), ll.Tags(), ll.Fields(), ll.Time())
	}
}

// AddError appends the given error to Accumulator.Errors.
func (a *Accumulator) AddError(err error) {
	if err == nil {
		return
	}
	a.Lock()
	a.Errors = append(a.Errors, err)
	if a.Cond != nil {
		a.Cond.Broadcast()
	}
	a.Unlock()
}

func (a *Accumulator) Debug() bool {
	// stub for implementing Accumulator interface.
	return a.debug
}

func (a *Accumulator) SetDebug(debug bool) {
	// stub for implementing Accumulator interface.
	a.debug = debug
}

//
// Utility functions
//

// GatherError calls the given Gather function and returns the first error found.
func (a *Accumulator) GatherError(gf func(optic.Accumulator) error) error {
	if err := gf(a); err != nil {
		return err
	}
	if len(a.Errors) > 0 {
		return a.Errors[0]
	}
	return nil
}

// FieldsCount returns the total number of fields in the accumulator, across all
// events.
func (a *Accumulator) FieldsCount() int {
	a.Lock()
	defer a.Unlock()
	counter := 0
	for _, e := range a.Events {
		for _ = range e.Fields {
			counter++
		}
	}
	return counter
}

// Wait waits for the given number of events to be added to the accumulator.
func (a *Accumulator) Wait(n int) {
	a.Lock()
	if a.Cond == nil {
		a.Cond = sync.NewCond(&a.Mutex)
	}
	for int(a.Count()) < n {
		a.Cond.Wait()
	}
	a.Unlock()
}

// WaitError waits for the given number of errors to be added to the accumulator.
func (a *Accumulator) WaitError(n int) {
	a.Lock()
	if a.Cond == nil {
		a.Cond = sync.NewCond(&a.Mutex)
	}
	for len(a.Errors) < n {
		a.Cond.Wait()
	}
	a.Unlock()
}

func (a *Accumulator) AssertContainsMetricWithTaggedFields(
	t *testing.T,
	name string,
	tags map[string]string,
	fields map[string]interface{},
) {
	a.Lock()
	defer a.Unlock()
	for _, e := range a.Events {
		if !reflect.DeepEqual(tags, e.Tags) {
			continue
		}

		if e.Name == name {
			assert.Equal(t, fields, e.Fields)
			return
		}
	}
	msg := fmt.Sprintf("unknown measurement %s with tags %v", name, tags)
	assert.Fail(t, msg)
}

// HasMetric returns true if the accumulator has a metric with the given
// name.
func (a *Accumulator) HasMetric(name string) bool {
	a.Lock()
	defer a.Unlock()
	for _, e := range a.Events {
		if e.Name == name {
			return true
		}
	}
	return false
}

// GetMetric gets the specified metric point from the accumulator.
func (a *Accumulator) GetMetric(name string) (*Event, bool) {
	for _, e := range a.Events {
		if e.Name == name {
			return e, true
		}
	}

	return nil, false
}

// Check the interfaces are satisfied
var (
	_ optic.Accumulator = &Accumulator{}
)
