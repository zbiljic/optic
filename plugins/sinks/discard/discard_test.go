package discard

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestDiscard_impl(t *testing.T) {
	var _ optic.Sink = new(Discard)
}
