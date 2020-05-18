package timer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConstantTimer(t *testing.T) {
	t.Run("with invalid duration", func(t *testing.T) {
		timer, err := NewConstantTimer(5 * time.Millisecond)
		assert.Nil(t, timer)
		assert.Error(t, err)
	})

	t.Run("with valid duration", func(t *testing.T) {
		timer, err := NewConstantTimer(5 * time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, timer)
		assert.Equal(t, 5*time.Second, timer.duration)
	})
}

func TestNext(t *testing.T) {
	timer, err := NewConstantTimer(5 * time.Second)
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, timer.Next(nil))
}

func TestReset(t *testing.T) {
	timer, err := NewConstantTimer(5 * time.Second)
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, timer.duration)
}
