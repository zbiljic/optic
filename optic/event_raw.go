package optic

import "time"

// A Raw is used for arbitrary bytes, with some optional metadata.
type Raw interface {
	Event

	// The source from which this event originated from.
	Source() string
	// The content received for the event.
	Value() []byte
}

type rawAccumulator interface {
	AddRaw(
		source string,
		value []byte,
		tags map[string]string,
		fields map[string]interface{},
		ts ...time.Time)
}
