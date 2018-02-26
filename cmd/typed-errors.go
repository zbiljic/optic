package cmd

import "github.com/pkg/errors"

var (
	errDummy = func() error {
		return errors.New("")
	}

	errInvalidCommandCall = func(cmdName string) error {
		return errors.Errorf("Run '%s help %s' for usage.", AppName, cmdName)
	}
)
