package memory

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestMemory_impl(t *testing.T) {
	var _ optic.Buffer = new(Memory)
}
