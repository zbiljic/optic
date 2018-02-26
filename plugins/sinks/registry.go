package sinks

import "github.com/zbiljic/optic/optic"

type Creator func() optic.Sink

var Sinks = map[string]Creator{}

func Add(name string, creator Creator) {
	Sinks[name] = creator
}
