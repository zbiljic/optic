package optic

type Sink interface {
	Plugin

	// Connect to the Sink.
	Connect() error

	// Close any connections to the Sink.
	Close() error

	// Write takes in group of events to be written to the Sink.
	Write(events []Event) error
}

type ServiceSink interface {
	Sink

	// Start the "service" that will provide an Sink.
	Start() error

	// Stop the "service" that will provide an Sink.
	Stop()
}
