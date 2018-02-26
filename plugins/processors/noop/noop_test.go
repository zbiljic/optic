package noop

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestNoop_impl(t *testing.T) {
	var _ optic.Processor = new(Noop)
}
