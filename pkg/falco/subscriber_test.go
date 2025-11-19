package falco

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubscriber(t *testing.T) {
	cfg := config.FalcoConfig{
		Enabled:  true,
		Hostname: "localhost",
		Port:     5060,
	}

	sub, err := NewSubscriber(cfg)
	require.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, cfg, sub.cfg)
}
