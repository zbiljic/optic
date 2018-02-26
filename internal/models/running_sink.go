package models

import (
	"log"
	"sync"
	"time"

	"github.com/zbiljic/pkg/metrics"

	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/optic"
)

const (
	// DefaultEventBatchSize is default size of events batch size.
	DefaultEventBatchSize = 1000
)

type RunningSink struct {
	Sink   optic.Sink
	Config *SinkConfig

	BufferSize    metrics.Gauge
	BufferLimit   metrics.Gauge
	EventsWritten metrics.Counter
	WriteTime     metrics.Histogram

	buffer optic.Buffer

	// Guards against concurrent calls to the Sink
	mu sync.Mutex
}

func NewRunningSink(sink optic.Sink, config *SinkConfig) *RunningSink {
	if config.EventBatchSize <= 0 {
		config.EventBatchSize = DefaultEventBatchSize
	}
	r := &RunningSink{
		Sink:   sink,
		Config: config,
		BufferSize: selfmetric.GetOrRegisterGauge(
			"sink",
			"buffer_size",
			map[string]string{"sink": config.Name},
		),
		BufferLimit: selfmetric.GetOrRegisterGauge(
			"sink",
			"buffer_limit",
			map[string]string{"sink": config.Name},
		),
		EventsWritten: selfmetric.GetOrRegisterCounter(
			"sink",
			"events_written",
			map[string]string{"sink": config.Name},
		),
		WriteTime: selfmetric.GetOrRegisterHistogram(
			"sink",
			"write_time_nanoseconds",
			map[string]string{"sink": config.Name},
		),
		buffer: config.Buffer,
	}

	r.BufferLimit.Update(int64(config.Buffer.Cap()))

	if r.Config.Encoder != nil {
		// configure encoder if possible
		if eo, ok := r.Sink.(optic.EncoderOutput); ok {
			eo.SetEncoder(r.Config.Encoder)
		}
	}

	return r
}

// SinkConfig containing a kind, name, and buffer.
type SinkConfig struct {
	Kind string
	Name string

	Buffer optic.Buffer

	EventBatchSize int

	Encoder optic.Encoder
}

func (r *RunningSink) Name() string {
	return "sinks." + r.Config.Name
}

// WriteEvent adds an event to the sink.
func (r *RunningSink) WriteEvent(event optic.Event) {
	if event == nil {
		return
	}

	r.mu.Lock()
	r.buffer.Append(event)
	r.mu.Unlock()

	if r.buffer.Len() >= r.Config.EventBatchSize {
		r.Write()
	}

}

// Write writes all cached events to this sink.
func (r *RunningSink) Write() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	nEvents := r.buffer.Len()
	r.BufferSize.Update(int64(nEvents))
	log.Printf("DEBUG Sink [%s] buffer fullness: %d / %d events. ",
		r.Config.Name, nEvents, r.buffer.Cap())

	var (
		batch []optic.Event
		err   error
	)

	start := 0
	end := r.Config.EventBatchSize

	i := 0
	for {
		if i > 0 && r.buffer.Len() < r.Config.EventBatchSize {
			break
		}
		i++

		batch, err = r.buffer.Slice(start, end)
		if err != nil {
			return err
		}

		err = r.write(batch)
		// If this fails, don't bother, just continue writing other events.
		// It will be removed once new events come in.
		if err == nil {
			r.buffer.RemoveRange(start, end)
			continue
		}

		// prepare for next iteration
		start += r.Config.EventBatchSize
		end += r.Config.EventBatchSize
		if start > nEvents {
			break
		}
	}

	return nil
}

func (r *RunningSink) write(events []optic.Event) error {
	count := len(events)
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	err := r.Sink.Write(events)
	elapsed := time.Since(start)

	if err == nil {
		log.Printf("DEBUG Sink [%s] wrote batch of %d events in %s",
			r.Config.Name, count, elapsed)
		r.EventsWritten.Inc(int64(count))
		r.WriteTime.Update(elapsed.Nanoseconds())
	}
	return err
}
