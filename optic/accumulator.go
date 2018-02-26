package optic

// Accumulator is an interface for "accumulating" events from source(s).
// The event is sent down a channel shared between plugins.
type Accumulator interface {
	AddEvent(Event)

	rawAccumulator
	metricAccumulator
	loglineAccumulator

	AddError(err error)
}
