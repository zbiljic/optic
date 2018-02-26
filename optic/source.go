package optic

type Source interface {
	Plugin

	// Gather takes in an accumulator and adds the telemetry that the Source
	// gathers. This is called every "interval".
	Gather(Accumulator) error
}

type ServiceSource interface {
	Source

	// Start starts the ServiceSource's service, whatever that may be.
	Start(Accumulator) error

	// Stop stops the services and closes any necessary channels and connections.
	Stop()
}
