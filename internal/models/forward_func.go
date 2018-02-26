package models

import (
	"log"

	"github.com/zbiljic/optic/optic"
)

// forwardFunc is used by both `RunningSource` and `RunningProcessor` to forward
// events down the pipeline.
func forwardFunc(
	name string,
	forwardProcessors []*RunningProcessor,
	forwardSinks []*RunningSink,
) func(optic.Event) {

	lenFilters := len(forwardProcessors)
	lenSinks := len(forwardSinks)
	lenForwards := lenFilters + lenSinks

	switch {
	case lenForwards == 1:
		if lenFilters == 1 {
			return func(event optic.Event) {
				forwardProcessors[0].ForwardEvent(event)
			}
		}
		return func(event optic.Event) {
			forwardSinks[0].WriteEvent(event)
		}
	case lenForwards > 1:
		switch {
		case lenFilters == 1 && lenSinks == 1:
			return func(event optic.Event) {
				forwardProcessors[0].ForwardEvent(event.Copy())
				forwardSinks[0].WriteEvent(event)
			}
		case lenFilters > 0 && lenSinks > 0:
			return func(event optic.Event) {
				for _, ff := range forwardProcessors {
					ff.ForwardEvent(event.Copy())
				}
				for i, fs := range forwardSinks {
					if i == lenSinks-1 {
						fs.WriteEvent(event)
					} else {
						fs.WriteEvent(event.Copy())
					}
				}
			}
		case lenFilters > 0:
			return func(event optic.Event) {
				for i, ff := range forwardProcessors {
					if i == lenFilters-1 {
						ff.ForwardEvent(event)
					} else {
						ff.ForwardEvent(event.Copy())
					}
				}
			}
		case lenSinks > 0:
			return func(event optic.Event) {
				for i, fs := range forwardSinks {
					if i == lenSinks-1 {
						fs.WriteEvent(event)
					} else {
						fs.WriteEvent(event.Copy())
					}
				}
			}
		}
	}

	log.Printf("INFO [%s] will not forward events anywhere.", name)

	return func(event optic.Event) {
		// no-op
	}
}
