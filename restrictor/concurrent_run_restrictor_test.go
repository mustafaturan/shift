package restrictor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConcurrentRunRestrictor(t *testing.T) {
	t.Run("valid options", func(t *testing.T) {
		r, err := NewConcurrentRunRestrictor("test", 2)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.IsType(t, &ConcurrentRunRestrictor{}, r)
		assert.Equal(t, "test", r.name)
		assert.Equal(t, int64(2), r.threshold)
		assert.Equal(t, int64(0), r.current)
	})

	t.Run("invalid options", func(t *testing.T) {
		r, err := NewConcurrentRunRestrictor("test", -1)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, r)
	})
}

func TestCheck(t *testing.T) {
	t.Run("under threshold", func(t *testing.T) {
		r, _ := NewConcurrentRunRestrictor("test", 1)
		ok, err := r.Check(context.Background())
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("over threshold", func(t *testing.T) {
		r, _ := NewConcurrentRunRestrictor("test", 1)

		go func() {
			_, err := r.Check(context.Background())
			require.NoError(t, err)
		}()

		time.Sleep(50 * time.Millisecond)
		ok, err := r.Check(context.Background())
		assert.False(t, ok)
		assert.Error(t, err)
		assert.IsType(t, &ThresholdError{}, err)
		assert.True(t, r.current > r.threshold)
	})
}

func TestDefer(t *testing.T) {
	r, _ := NewConcurrentRunRestrictor("test", 1)
	_, err := r.Check(context.Background())
	require.NoError(t, err)

	assert.Equal(t, int64(1), r.current)
	r.Defer()
	assert.Equal(t, int64(0), r.current)
}
