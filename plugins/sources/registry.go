package sources

import "github.com/zbiljic/optic/optic"

type Creator func() optic.Source

var Sources = map[string]Creator{}

func Add(name string, creator Creator) {
	Sources[name] = creator
}
