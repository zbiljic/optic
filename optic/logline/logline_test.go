package logline

import (
	"testing"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestLogLine_impl(t *testing.T) {
	var _ optic.LogLine = new(logline)
}
