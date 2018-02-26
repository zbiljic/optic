package raw

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestRaw_impl(t *testing.T) {
	var _ optic.Raw = new(raw)
}
