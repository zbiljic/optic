package logline

import (
	"fmt"
	"time"

	"github.com/zbiljic/optic/optic"
)

type logline struct {
	ts      time.Time
	tags    map[string]string
	fields  map[string]interface{}
	path    string
	content string
}

func New(
	path string,
	content string,
	tags map[string]string,
	fields map[string]interface{},
	t ...time.Time,
) (optic.LogLine, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("missing logline content")
	}

	var ts time.Time
	if len(t) > 0 {
		ts = t[0]
	} else {
		ts = time.Now()
	}

	ll := &logline{
		ts:      ts,
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
		path:    path,
		content: content,
	}

	if tags != nil {
		for k, v := range tags {
			if len(k) == 0 {
				continue
			}
			ll.tags[k] = v
		}
	}

	if fields != nil {
		for k, v := range fields {
			ll.fields[k] = v
		}
	}

	return ll, nil
}

func (l *logline) Type() optic.EventType {
	return optic.LogLineEvent
}

func (l *logline) Time() time.Time {
	return l.ts
}

func (l *logline) Tags() map[string]string {
	return l.tags
}

func (l *logline) HasTag(key string) bool {
	_, ok := l.tags[key]
	return ok
}

func (l *logline) AddTag(key string, value string) {
	l.tags[key] = value
}

func (l *logline) RemoveTag(key string) {
	delete(l.tags, key)
}

func (l *logline) Fields() map[string]interface{} {
	return l.fields
}

func (l *logline) HasField(key string) bool {
	_, ok := l.fields[key]
	return ok
}

func (l *logline) AddField(key string, value interface{}) {
	l.fields[key] = value
}

func (l *logline) RemoveField(key string) {
	delete(l.fields, key)
}

func (l *logline) Serialize() []byte {
	return []byte(l.content)
}

func (l *logline) String() string {
	return l.content
}

func (l *logline) Copy() optic.Event {
	return copyFrom(l)
}

func (l *logline) Path() string {
	return l.path
}

func (l *logline) Content() string {
	return l.content
}

func copyFrom(l *logline) optic.LogLine {
	out := logline{
		ts:      l.ts,
		tags:    make(map[string]string),
		fields:  make(map[string]interface{}),
		path:    l.path,
		content: l.content,
	}
	for k, v := range l.tags {
		out.tags[k] = v
	}
	for k, v := range l.fields {
		out.fields[k] = v
	}
	return &out
}
