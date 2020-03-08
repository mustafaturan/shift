package shift

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/mustafaturan/shift/timers"
	"github.com/stretchr/testify/assert"
)

func TestRun_WithRestrictionCheck(t *testing.T) {
	t.Run("allowed", func(t *testing.T) {
		restrictor := &fakeRestrictor{res: true, err: nil}
		var onFailureHandler OnFailure = func(State, error) {}
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnFailureHandlers(onFailureHandler),
			WithRestrictors(restrictor),
		)
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return "TestRun_WithRestrictionCheck", nil
		}
		res, err := cb.Run(ctx, fn)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("disallowed", func(t *testing.T) {
		restrictor := &fakeRestrictor{res: false, err: errors.New("fake")}
		var onFailureHandler OnFailure = func(State, error) {}
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnFailureHandlers(onFailureHandler),
			WithRestrictors(restrictor),
		)
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return "TestRun_WithRestrictionCheck", nil
		}
		res, err := cb.Run(ctx, fn)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestRun_OnStateClose(t *testing.T) {
	var onSuccessHandler OnSuccess = func(_ interface{}) {}
	var onFailureHandler OnFailure = func(State, error) {}
	var onStateChangeHandler OnStateChange = func(State, State) {}
	t.Run("on success", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnSuccessHandlers(onSuccessHandler),
			WithInitialState(StateClose),
		)
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return "TestRun_OnStateClose", nil
		}
		res, err := cb.Run(ctx, fn)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("on failure", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnFailureHandlers(onFailureHandler),
			WithInitialState(StateClose),
		)
		cb.failureThreshold = 2
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return nil, errors.New("foo")
		}
		res, err := cb.Run(ctx, fn)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("on failure threshold", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnFailureHandlers(onFailureHandler),
			WithOnStateChangeHandlers(onStateChangeHandler),
			WithInitialState(StateClose),
		)
		cb.failureThreshold = 1
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return nil, errors.New("foo")
		}
		res, err := cb.Run(ctx, fn)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.True(t, cb.State().isOpen())
	})
}

func TestRun_OnStateHalfOpen(t *testing.T) {
	var onSuccessHandler OnSuccess = func(_ interface{}) {}
	var onFailureHandler OnFailure = func(State, error) {}
	var onStateChangeHandler OnStateChange = func(State, State) {}
	t.Run("on success", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnSuccessHandlers(onSuccessHandler),
			WithInitialState(StateHalfOpen),
		)
		cb.successThreshold = 3
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return "TestRun_OnStateHalfOpen", nil
		}
		res, err := cb.Run(ctx, fn)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, cb.State().isHalfOpen())
	})

	t.Run("on success threshold", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnSuccessHandlers(onSuccessHandler),
			WithInitialState(StateHalfOpen),
		)
		cb.failure = 3
		cb.successThreshold = 1
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return "TestRun_OnStateHalfOpen", nil
		}
		res, err := cb.Run(ctx, fn)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.True(t, cb.State().isClose())
		assert.Equal(t, int32(0), cb.failure)
	})

	t.Run("on failure", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithOnFailureHandlers(onFailureHandler),
			WithOnStateChangeHandlers(onStateChangeHandler),
			WithInitialState(StateHalfOpen),
		)
		ctx := context.Background()
		now := time.Now()
		cb.resetAt = now
		cb.failureThreshold = 1
		var fn Operate = func(context.Context) (interface{}, error) {
			return nil, errors.New("foo")
		}
		res, err := cb.Run(ctx, fn)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.True(t, cb.State().isOpen())
		assert.True(t, cb.resetAt.Sub(now) > 0)
	})
}

func TestRun_OnStateOpen(t *testing.T) {
	var onFailureHandler OnFailure = func(State, error) {}
	cb, _ := NewCircuitBreaker(
		"test",
		WithOnFailureHandlers(onFailureHandler),
		WithInitialState(StateOpen),
	)
	ctx := context.Background()
	var fn Operate = func(context.Context) (interface{}, error) {
		return "TestRun_OnStateOpen", nil
	}
	res, err := cb.Run(ctx, fn)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.True(t, cb.State().isOpen())
}

func TestRun_TryToClose_OnClose(t *testing.T) {
	cb, _ := NewCircuitBreaker("test", WithInitialState(StateClose))

	s, ok := cb.tryToClose()
	assert.False(t, ok)
	assert.True(t, s.isClose())
}

func TestRun_TryToHalfOpen_OnHalfOpen(t *testing.T) {
	cb, _ := NewCircuitBreaker("test", WithInitialState(StateHalfOpen))

	s, ok := cb.tryToHalfOpen()
	assert.False(t, ok)
	assert.True(t, s.isHalfOpen())
}

func TestRun_TryToOpen_OnOpen(t *testing.T) {
	cb, _ := NewCircuitBreaker("test", WithInitialState(StateOpen))

	s, ok := cb.tryToOpen()
	assert.False(t, ok)
	assert.True(t, s.isOpen())
}

func TestRun_WithTimeoutError(t *testing.T) {
	t.Run("with timeout", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithInvocationTimeout(28*time.Millisecond),
			WithInitialState(StateClose),
		)
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return "TestRun_WithTimeoutError", nil
		}

		s, err := cb.Run(ctx, fn)
		assert.Error(t, err)
		assert.IsType(t, &TimeoutError{}, err)
		assert.Nil(t, s)
		time.Sleep(2 * time.Second)
	})

	t.Run("without timeout", func(t *testing.T) {
		cb, _ := NewCircuitBreaker(
			"test",
			WithInvocationTimeout(200*time.Millisecond),
			WithInitialState(StateClose),
		)
		ctx := context.Background()
		var fn Operate = func(context.Context) (interface{}, error) {
			return "TestRun_WithTimeoutError", nil
		}

		s, err := cb.Run(ctx, fn)
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})
}

func TestState(t *testing.T) {
	tests := []struct {
		state State
	}{
		{state: StateClose},
		{state: StateHalfOpen},
		{state: StateOpen},
	}

	for _, test := range tests {
		cb, _ := NewCircuitBreaker("test", WithInitialState(test.state))
		assert.Equal(t, test.state, cb.State())
	}
}

func TestOverride(t *testing.T) {
	tests := []struct {
		state State
	}{
		{state: StateClose},
		{state: StateHalfOpen},
		{state: StateOpen},
	}

	for _, test := range tests {
		cb, _ := NewCircuitBreaker("test")
		cb.Override(test.state)
		assert.Equal(t, test.state, cb.State())
	}
}

func TestOverride_StateOpenWithResetTimeout(t *testing.T) {
	timer := timers.NewConstantTimer(50 * time.Millisecond)
	cb, _ := NewCircuitBreaker(
		"test",
		WithResetTimer(timer),
	)
	cb.Override(StateOpen)
	assert.Equal(t, StateOpen, cb.State())
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, StateHalfOpen, cb.State())
}

func TestCircuitBreakerOverrideError(t *testing.T) {
	err := &CircuitBreakerOverrideError{Name: "test"}
	assert.EqualError(t, err, "circuit breaker(test) is overridden as open")
}

func TestCircuitBreakerIsOpenError(t *testing.T) {
	now := time.Now()
	err := &CircuitBreakerIsOpenError{Name: "test", ExpiresAt: now}
	assert.EqualError(t, err, fmt.Sprintf("circuit breaker(test) is open, will transition to half-open at %+v", now))
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{Name: "test", Duration: 2 * time.Second}
	assert.EqualError(t, err, "invocation timeout for test circuit breaker on 2s")
}

type fakeRestrictor struct {
	res bool
	err error
}

func (r *fakeRestrictor) Check() (bool, error) { return r.res, r.err }
func (r *fakeRestrictor) Defer()               {}
