package printer

import (
	"fmt"

	"github.com/zbiljic/optic/optic"
	"github.com/zbiljic/optic/plugins/processors"
)

const (
	name        = "printer"
	description = `Print all events that pass through this processor.`
)

type Printer struct{}

func NewPrinter() optic.Processor {
	return &Printer{}
}

func (*Printer) Kind() string {
	return name
}

func (*Printer) Description() string {
	return description
}

func (*Printer) Init() error {
	return nil
}

func (p *Printer) Apply(in ...optic.Event) []optic.Event {
	for _, event := range in {
		fmt.Println(event.String())
	}
	return in
}

func init() {
	processors.Add(name, NewPrinter)
}
