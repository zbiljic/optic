package metric

import (
	"fmt"
	"time"

	"github.com/zbiljic/optic/optic"
)

type metric struct {
	ts         time.Time
	tags       map[string]string
	fields     map[string]interface{}
	name       string
	metricType optic.MetricType
}

func New(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
	metricType ...optic.MetricType,
) (optic.Metric, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("missing metric name")
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("%s: must have one or more fields", name)
	}
	if !MetricNameValid(name) {
		return nil, fmt.Errorf("invalid metric name: %s", name)
	}

	var thisType optic.MetricType
	if len(metricType) > 0 {
		thisType = metricType[0]
	} else {
		thisType = optic.UntypedMetric
	}

	m := &metric{
		ts:         t,
		tags:       make(map[string]string),
		fields:     make(map[string]interface{}),
		name:       sanitize(name, "name"),
		metricType: thisType,
	}

	// process tags map
	for k, v := range tags {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		m.tags[sanitize(k, "tagkey")] = v
	}

	// process fields map
	for k := range fields {
		if !MetricNameValid(k) {
			return nil, fmt.Errorf("%s: invalid field key: %s", name, k)
		}
	}

	for k, v := range fields {
		m.fields[sanitize(k, "fieldkey")] = v
	}

	return m, nil
}

func NewParsed(
	name string,
	tags map[string]string,
	fields map[string]interface{},
	t time.Time,
	metricType ...optic.MetricType,
) (optic.Metric, error) {

	var thisType optic.MetricType
	if len(metricType) > 0 {
		thisType = metricType[0]
	} else {
		thisType = optic.UntypedMetric
	}

	m := &metric{
		ts:         t,
		tags:       tags,
		fields:     fields,
		name:       name,
		metricType: thisType,
	}

	return m, nil
}

func (*metric) Type() optic.EventType {
	return optic.MetricEvent
}

func (m *metric) Time() time.Time {
	return m.ts
}

func (m *metric) Tags() map[string]string {
	return m.tags
}

func (m *metric) HasTag(key string) bool {
	_, ok := m.tags[sanitize(key, "tagkey")]
	return ok
}

func (m *metric) AddTag(key string, value string) {
	m.tags[sanitize(key, "tagkey")] = value
}

func (m *metric) RemoveTag(key string) {
	delete(m.tags, sanitize(key, "tagkey"))
}

func (m *metric) Fields() map[string]interface{} {
	return m.fields
}

func (m *metric) HasField(key string) bool {
	_, ok := m.fields[sanitize(key, "fieldkey")]
	return ok
}

func (m *metric) AddField(key string, value interface{}) {
	m.fields[sanitize(key, "fieldkey")] = value
}

func (m *metric) RemoveField(key string) {
	delete(m.fields, sanitize(key, "fieldkey"))
}

func (m *metric) Serialize() []byte {
	return serialize(m)
}

func (m *metric) String() string {
	b := serialize(m)
	return string(b)
}

func (m *metric) Copy() optic.Event {
	return copyFrom(m)
}

func (m *metric) Name() string {
	return m.name
}

func (m *metric) MetricType() optic.MetricType {
	return m.metricType
}

func (m *metric) SetName(name string) {
	m.name = sanitize(name, "name")
}

func (m *metric) SetPrefix(prefix string) {
	m.name = namePrefixSanitizer.ReplaceAllString(prefix, "_") + m.name
}

func (m *metric) SetSuffix(suffix string) {
	m.name = m.name + nameBodySanitizer.ReplaceAllString(suffix, "_")
}

func (m *metric) Split() []optic.Metric {
	if len(m.fields) == 1 {
		return []optic.Metric{m}
	}
	var out []optic.Metric

	for k, v := range m.fields {
		out = append(out, splitCopy(m, k, v))
	}

	return out
}

func copyFrom(m *metric) optic.Metric {
	out := metric{
		ts:         m.ts,
		tags:       make(map[string]string),
		fields:     make(map[string]interface{}),
		name:       m.name,
		metricType: m.metricType,
	}
	for k, v := range m.tags {
		out.tags[k] = v
	}
	for k, v := range m.fields {
		out.fields[k] = v
	}
	return &out
}

func splitCopy(m *metric, fieldsKey string, fieldValue interface{}) optic.Metric {
	out := metric{
		ts:         m.ts,
		tags:       make(map[string]string),
		fields:     map[string]interface{}{fieldsKey: fieldValue},
		name:       m.name,
		metricType: m.metricType,
	}
	for k, v := range m.tags {
		out.tags[k] = v
	}
	return &out
}
