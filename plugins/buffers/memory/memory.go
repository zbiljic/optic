package memory

import (
	"fmt"

	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/plugins/buffers"
)

const (
	name        = "memory"
	description = `Memory is a buffer which stores events in memory.`
)

const (
	// DefaultBufferLimit represents default size of memory buffer.
	DefaultBufferLimit = 1000
)

type Memory struct {
	// Limit is the maximum number of events that the buffer can store.
	Limit int `mapstructure:"limit"`

	buffer []optic.Event
}

func NewMemory() optic.Buffer {
	return &Memory{
		Limit: DefaultBufferLimit,
	}
}

func (*Memory) Kind() string {
	return name
}

func (*Memory) Description() string {
	return description
}

func (m *Memory) Build() error {
	if m.Limit <= 0 {
		return fmt.Errorf("Buffer limit must be positive number: %d", m.Limit)
	}
	m.buffer = make([]optic.Event, 0)
	return nil
}

func (m *Memory) Len() int {
	return len(m.buffer)
}

func (m *Memory) Cap() int {
	return m.Limit
}

func (m *Memory) Append(events ...optic.Event) {

	over := m.Limit - len(m.buffer) - len(events)
	if over < 0 {
		// not enough place for new events

		if len(events) > m.Limit {
			// SPECIAL CASE: new events will overwrite all elements
			// only write last ones

			// index for events array
			ei := len(events) - m.Limit

			for i := range m.buffer {
				// we must clear first
				m.buffer[i] = nil
				m.buffer[i] = events[ei+i]
			}
			return
		}

		m.RemoveRange(0, -over)
	}

	m.buffer = append(m.buffer, events...)
}

func (m *Memory) Slice(start int, end int) ([]optic.Event, error) {
	if start < 0 || end <= start || start >= len(m.buffer) {
		return []optic.Event{}, nil
	}
	if end > len(m.buffer) {
		end = len(m.buffer)
	}

	return m.buffer[start:end], nil
}

func (m *Memory) RemoveRange(from int, to int) {
	if from < 0 || to <= from || from >= len(m.buffer) {
		return
	}
	if to > len(m.buffer) {
		to = len(m.buffer)
	}

	copy(m.buffer[from:], m.buffer[to:])
	n := len(m.buffer)
	l := n - to + from
	for i := l; i < n; i++ {
		m.buffer[i] = nil
	}
	m.buffer = m.buffer[:l]
}

func (m *Memory) Clear() {
	for i := range m.buffer {
		m.buffer[i] = nil
	}
	m.buffer = m.buffer[:0]
}

func (m *Memory) IsEmpty() bool {
	return len(m.buffer) == 0
}

func (m *Memory) Close() error {
	return nil
}

func init() {
	buffers.Add(name, NewMemory)
}
