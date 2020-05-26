package shift

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvalidOptionError(t *testing.T) {
	err := &InvalidOptionError{
		Name:    "test",
		Message: "can't be nil",
	}
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid option provided for test: can't be nil")
}

func TestUnknownStateError(t *testing.T) {
	err := &UnknownStateError{
		State: State(int8(-1)),
	}
	assert.Error(t, err)
	assert.EqualError(t, err, "unknown state(-1) provided, the allowed states are 'close', 'half-open' and 'open'")
}

func TestIsAlreadyInDesiredStateError(t *testing.T) {
	err := &IsAlreadyInDesiredStateError{
		Name:  "test",
		State: StateOpen,
	}

	assert.Error(t, err)
	assert.EqualError(t, err, "circuit breaker(test) is already in the desired state(open)")
}

func TestIsOnOpenStateError(t *testing.T) {
	err := &IsOnOpenStateError{}

	assert.Error(t, err)
	assert.EqualError(t, err, "is on open state")
}

func TestInvocationError(t *testing.T) {
	err := &InvocationError{
		Name: "test",
		Err: &InvocationTimeoutError{
			Duration: 5 * time.Second,
		},
	}

	assert.Error(t, err)
	assert.EqualError(t, err, "circuit breaker(test) invocation failed with invocation timeout on 5s")
	assert.EqualError(t, errors.Unwrap(err), "invocation timeout on 5s")
}

func TestInvocationTimeoutError(t *testing.T) {
	err := &InvocationTimeoutError{
		Duration: 5 * time.Second,
	}

	assert.Error(t, err)
	assert.EqualError(t, err, "invocation timeout on 5s")
}

func TestFailureThresholdReachedError(t *testing.T) {
	err := &FailureThresholdReachedError{}

	assert.Error(t, err)
	assert.EqualError(t, err, "failure threshold reached")
}
