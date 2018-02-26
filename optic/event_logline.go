package optic

import "time"

// A LogLine represents unstructured piece of text, plus associated metadata.
type LogLine interface {
	Event

	// The path that this `LogLine` originated from.
	Path() string
	// The line read from the `LogLine` path.
	Content() string
}

type loglineAccumulator interface {
	// AddLogLine adds a log line to the accumulator with the given path, content,
	// tags and fields (and timestamp). If a timestamp is not provided, then the
	// accumulator sets it to "now".
	//
	// NOTE: tags and fields are expected to be owned by the caller, don't mutate
	// it after passing to `AddLogLine`.
	AddLogLine(
		path string,
		content string,
		tags map[string]string,
		fields map[string]interface{},
		ts ...time.Time)
}
