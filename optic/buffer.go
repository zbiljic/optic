package optic

// Buffer represents a method of storing event messages.
type Buffer interface {
	Plugin

	// Build builds the buffer based on the configuration.
	Build() error

	// Len returns the current number of events in the buffer.
	Len() int

	// Cap returns the maximum capacity of the buffer.
	Cap() int

	// Append will add the specified events in the buffer.
	Append(events ...Event)

	// Slice is used for retreiving events from buffer. If there are some items
	// in the buffer in the requested range it will return them.
	// It WILL NOT fail if the offsets ara out of range, it will just return
	// empty array.
	Slice(start, end int) ([]Event, error)

	// RemoveRange removes all of the elements whose index is between from,
	// inclusive, and to, exclusive, from this buffer.
	RemoveRange(from, to int)

	// Clear removes all elements from the buffer.
	Clear()

	// IsEmpty returns true if Buffer is empty.
	IsEmpty() bool

	// Close closes the Buffer gracefully.
	Close() error
}
