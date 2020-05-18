package shift

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOnFailure(t *testing.T) {
	// Ensure OnFailure implements FailureHandler on build
	var _ FailureHandler = (OnFailure)(nil)

	var called bool
	var fn OnFailure = func(context.Context, error) {
		called = true
	}

	fn.Handle(context.Background(), nil)
	assert.Equal(t, true, called)
}

func TestOnSuccess(t *testing.T) {
	// Ensure OnSuccess implements SuccessHandler on build
	var _ SuccessHandler = (OnSuccess)(nil)

	var called bool
	var fn OnSuccess = func(context.Context, interface{}) {
		called = true
	}

	fn.Handle(context.Background(), nil)
	assert.Equal(t, true, called)
}

func TestOnStateChange(t *testing.T) {
	// Ensure OnStateChange implements StateChangeHandler on build
	var _ StateChangeHandler = (OnStateChange)(nil)

	var called bool
	var fn OnStateChange = func(State, State, Stats) {
		called = true
	}

	fn.Handle(StateClose, StateOpen, Stats{})
	assert.Equal(t, true, called)
}
