package optic

import "time"

// EventType is an enumeration of event types.
type EventType uint8

// Possible values for the EventType enum.
const (
	_ EventType = iota
	// A event for `Raw`. See its documentation for more detail.
	RawEvent
	// A event for `Metric`. See its documentation for more detail.
	MetricEvent
	// A event for `LogLine`. See its documentation for more detail.
	LogLineEvent
)

var eventTypes = map[EventType]string{
	RawEvent:     "raw",
	MetricEvent:  "metric",
	LogLineEvent: "logline",
}

func (et EventType) String() string {
	return eventTypes[et]
}

// Event is the central Optic datastructure.
//
// Event is the heart of Optic, that works in all cases. This datastructure goes
// through source / processor / sink operations depending on their
// implementation.
type Event interface {
	Type() EventType
	Time() time.Time
	Tags() map[string]string
	Fields() map[string]interface{}

	HasTag(key string) bool
	AddTag(key, value string)
	RemoveTag(key string)

	HasField(key string) bool
	AddField(key string, value interface{})
	RemoveField(key string)

	// Serialize serializes the event into a line-protocol byte buffer.
	Serialize() []byte
	// String serializes the event into a string.
	String() string
	// Copy deep-copies the event.
	Copy() Event
}
