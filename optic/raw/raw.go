package raw

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/zbiljic/optic/optic"
)

type raw struct {
	ts     time.Time
	tags   map[string]string
	fields map[string]interface{}
	source string
	value  []byte
}

func New(
	source string,
	value []byte,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) (optic.Raw, error) {
	if len(source) == 0 {
		return nil, fmt.Errorf("missing raw source")
	}

	v := make([]byte, len(value))
	copy(v, value)

	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now()
	}

	r := &raw{
		ts:     ts,
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
		source: source,
		value:  v,
	}

	if tags != nil {
		for k, v := range tags {
			if len(k) == 0 {
				continue
			}
			r.tags[k] = v
		}
	}

	if fields != nil {
		for k, v := range fields {
			r.fields[k] = v
		}
	}

	return r, nil
}

func (r *raw) Type() optic.EventType {
	return optic.RawEvent
}

func (r *raw) Time() time.Time {
	return r.ts
}

func (r *raw) Tags() map[string]string {
	return r.tags
}

func (r *raw) HasTag(key string) bool {
	_, ok := r.tags[key]
	return ok
}

func (r *raw) AddTag(key string, value string) {
	r.tags[key] = value
}

func (r *raw) RemoveTag(key string) {
	delete(r.tags, key)
}

func (r *raw) Fields() map[string]interface{} {
	return r.fields
}

func (r *raw) HasField(key string) bool {
	_, ok := r.fields[key]
	return ok
}

func (r *raw) AddField(key string, value interface{}) {
	r.fields[key] = value
}

func (r *raw) RemoveField(key string) {
	delete(r.fields, key)
}

func (r *raw) Serialize() []byte {
	return r.serializeSimple()
}

func (r *raw) String() string {
	b := r.serializeSimple()
	return string(b)
}

func (r *raw) Copy() optic.Event {
	return copyFrom(r)
}

func (r *raw) Source() string {
	return r.source
}

func (r *raw) Value() []byte {
	return r.value
}

func copyFrom(r *raw) optic.Raw {
	out := raw{
		ts:     r.ts,
		tags:   make(map[string]string),
		fields: make(map[string]interface{}),
		source: r.source,
		value:  make([]byte, len(r.value)),
	}
	copy(out.value, r.value)
	for k, v := range r.tags {
		out.tags[k] = v
	}
	for k, v := range r.fields {
		out.fields[k] = v
	}
	return &out
}

func (r *raw) serializeSimple() []byte {
	if len(r.value) > 0 {
		return r.value
	}
	rm := map[string]interface{}{
		"timestamp": r.ts.UTC().Format(time.RFC3339Nano),
		"source":    r.source,
		"value":     r.value,
		"tags":      r.tags,
		"fields":    r.fields,
	}
	json, err := json.Marshal(rm)
	if err != nil {
		return []byte{}
	}
	return json
}
