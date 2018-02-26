package metric

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/zbiljic/optic/optic"
)

// Check the interfaces are satisfied
func TestMetric_impl(t *testing.T) {
	var _ optic.Metric = new(metric)
}

func TestNewMetric(t *testing.T) {
	now := time.Now()

	tags := map[string]string{
		"host":       "localhost",
		"datacenter": "us-east-1",
	}
	fields := map[string]interface{}{
		"usage_idle": float64(99),
		"usage_busy": float64(1),
	}
	m, err := New("cpu", tags, fields, now)
	assert.NoError(t, err)

	assert.Equal(t, optic.MetricEvent, m.Type())
	assert.Equal(t, now.UnixNano(), m.Time().UnixNano())
	assert.Equal(t, tags, m.Tags())
	assert.Equal(t, fields, m.Fields())
	assert.Equal(t, "cpu", m.Name())
	assert.Equal(t, optic.UntypedMetric, m.MetricType())
}
