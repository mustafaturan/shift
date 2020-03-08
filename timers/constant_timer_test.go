package timers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConstantTimer(t *testing.T) {
	timer := NewConstantTimer(5 * time.Second)
	assert.Equal(t, 5*time.Second, timer.duration)
}

func TestNext(t *testing.T) {
	timer := NewConstantTimer(5 * time.Second)
	assert.Equal(t, 5*time.Second, timer.Next(nil))
}

func TestReset(t *testing.T) {
	timer := NewConstantTimer(5 * time.Second)
	assert.Equal(t, 5*time.Second, timer.duration)
}
