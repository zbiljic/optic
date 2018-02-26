package line

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestLineCodec_impl(t *testing.T) {
	var _ optic.Codec = new(LineCodec)
}
