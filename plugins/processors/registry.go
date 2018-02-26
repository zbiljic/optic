package processors

import "github.com/zbiljic/optic/optic"

type Creator func() optic.Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	Processors[name] = creator
}
