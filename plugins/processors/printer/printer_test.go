package printer

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestPrinter_impl(t *testing.T) {
	var _ optic.Processor = new(Printer)
}
