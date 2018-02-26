package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zbiljic/optic/internal/config"
)

func TestAgent_OmitHostname(t *testing.T) {
	c := config.NewConfig()
	c.Agent.OmitHostname = true
	_, err := NewAgent(c)
	assert.NoError(t, err)
	assert.NotContains(t, c.Tags, "host")
}
