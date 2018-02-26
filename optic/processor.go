package optic

// Processor plugins are the mechanisms for transforming in-flight events.
type Processor interface {
	Plugin

	// Init executes any possible initialization code of the processor.
	Init() error

	// Apply the processor to the given event.
	Apply(in ...Event) []Event
}
