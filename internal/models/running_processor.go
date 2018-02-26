package models

import (
	"github.com/zbiljic/pkg/metrics"

	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/optic"
)

type RunningProcessor struct {
	Processor optic.Processor
	Config    *ProcessorConfig

	EventsProcessed metrics.Counter
	EventsFiltered  metrics.Counter

	forwardFunc func(optic.Event)
}

func NewRunningProcessor(processor optic.Processor, config *ProcessorConfig) *RunningProcessor {
	r := &RunningProcessor{
		Processor: processor,
		Config:    config,
		EventsProcessed: selfmetric.GetOrRegisterCounter(
			"processor",
			"events_processed",
			map[string]string{"processor": config.Name},
		),
		EventsFiltered: selfmetric.GetOrRegisterCounter(
			"processor",
			"events_filtered",
			map[string]string{"processor": config.Name},
		),
	}

	r.forwardFunc = forwardFunc(
		r.Name(),
		r.Config.ForwardProcessors,
		r.Config.ForwardSinks,
	)

	return r
}

// ProcessorConfig containing a kind, name.
type ProcessorConfig struct {
	Kind string
	Name string

	ForwardProcessors []*RunningProcessor
	ForwardSinks      []*RunningSink
}

func (r *RunningProcessor) Name() string {
	return "processors." + r.Config.Name
}

func (r *RunningProcessor) Apply(in ...optic.Event) []optic.Event {
	out := r.Processor.Apply(in...)
	diff := len(in) - len(out)
	r.EventsProcessed.Inc(int64(len(in)))
	r.EventsFiltered.Inc(int64(diff))
	return out
}

func (r *RunningProcessor) Flush() {
	// apply self with empty array
	events := r.Apply([]optic.Event{}...)

	switch len(events) {
	case 0:
		for _, ff := range r.Config.ForwardProcessors {
			ff.Flush()
		}
	case 1:
		r.forwardFunc(events[0])
	default:
		for _, event := range events {
			r.forwardFunc(event)
		}
	}
}

// ForwardEvent adds an event to the processor to be forwarded.
func (r *RunningProcessor) ForwardEvent(event optic.Event) {
	if event == nil {
		return
	}

	// apply self
	events := r.Apply(event)

	switch len(events) {
	case 0:
		return
	case 1:
		r.forwardFunc(events[0])
	default:
		for _, event := range events {
			r.forwardFunc(event)
		}
	}
}
