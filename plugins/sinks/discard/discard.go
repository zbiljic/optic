package discard

import (
	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/plugins/sinks"
)

const (
	name        = "discard"
	description = `Discards all received events.`
)

type Discard struct{}

func NewDiscard() optic.Sink {
	return &Discard{}
}

func (*Discard) Kind() string {
	return name
}

func (*Discard) Description() string {
	return description
}

func (d *Discard) Connect() error {
	return nil
}

func (d *Discard) Close() error {
	return nil
}

func (d *Discard) Write(events []optic.Event) error {
	return nil
}

func init() {
	sinks.Add(name, NewDiscard)
}
