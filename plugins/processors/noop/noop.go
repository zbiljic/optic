package noop

import (
	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/plugins/processors"
)

const (
	name        = "noop"
	description = `Noop is a no-op processor that does nothing, the events pass through unchanged.`
)

type Noop struct{}

func NewNoop() optic.Processor {
	return &Noop{}
}

func (*Noop) Kind() string {
	return name
}

func (*Noop) Description() string {
	return description
}

func (*Noop) Init() error {
	return nil
}

func (p *Noop) Apply(in ...optic.Event) []optic.Event {
	return in
}

func init() {
	processors.Add(name, NewNoop)
}
