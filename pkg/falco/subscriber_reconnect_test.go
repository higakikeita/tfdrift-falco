package falco

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/require"
)

// TestRunWithReconnect_RetriesUntilContextCancel proves the subscriber no
// longer dies on the first stream error (pus #9): a repeatedly-failing attempt
// is retried, and the loop exits only when the context is cancelled.
func TestRunWithReconnect_RetriesUntilContextCancel(t *testing.T) {
	s := &Subscriber{initialBackoff: time.Millisecond, maxBackoff: 2 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())

	var attempts int32
	attempt := func(_ context.Context, _ chan<- types.Event) error {
		n := atomic.AddInt32(&attempts, 1)
		s.connected.Store(true) // simulate an established stream
		if n >= 3 {
			cancel() // stop after a few reconnects
		}
		return errors.New("stream broke")
	}

	err := s.runWithReconnect(ctx, nil, attempt)

	require.ErrorIs(t, err, context.Canceled)
	require.GreaterOrEqual(t, atomic.LoadInt32(&attempts), int32(3), "must have retried, not died on first error")
	require.False(t, s.Connected(), "subscriber must report disconnected after the loop ends")
}

// TestRunWithReconnect_ReturnsImmediatelyIfCancelledBeforeStart guards the
// clean-shutdown path.
func TestRunWithReconnect_ReturnsImmediatelyIfCancelledBeforeStart(t *testing.T) {
	s := &Subscriber{initialBackoff: time.Millisecond, maxBackoff: time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var attempts int32
	attempt := func(_ context.Context, _ chan<- types.Event) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("should not be called")
	}

	err := s.runWithReconnect(ctx, nil, attempt)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, int32(0), atomic.LoadInt32(&attempts), "must not attempt when already cancelled")
}

// TestRunWithReconnect_StopsDuringBackoff verifies cancellation while waiting
// to reconnect returns promptly rather than sleeping out the backoff.
func TestRunWithReconnect_StopsDuringBackoff(t *testing.T) {
	s := &Subscriber{initialBackoff: time.Hour, maxBackoff: time.Hour} // long backoff
	ctx, cancel := context.WithCancel(context.Background())

	attempt := func(_ context.Context, _ chan<- types.Event) error {
		// fail once, then the loop enters the (very long) backoff wait
		go func() { time.Sleep(20 * time.Millisecond); cancel() }()
		return errors.New("boom")
	}

	done := make(chan error, 1)
	go func() { done <- s.runWithReconnect(ctx, nil, attempt) }()

	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("runWithReconnect did not return promptly on cancel during backoff")
	}
}
