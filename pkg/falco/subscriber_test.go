package falco

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
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

func TestStart_ContextCancellation(t *testing.T) {
	cfg := config.FalcoConfig{
		Enabled:    true,
		Hostname:   "localhost",
		Port:       5060,
		CertFile:   "",
		KeyFile:    "",
		CARootFile: "",
	}

	subscriber, err := NewSubscriber(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	eventCh := make(chan types.Event, 10)

	// Cancel immediately
	cancel()

	// Start should return context.Canceled error
	err = subscriber.Start(ctx, eventCh)
	assert.Error(t, err)
}
