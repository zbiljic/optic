package file

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestFile_impl(t *testing.T) {
	var _ optic.Sink = new(File)
}
