package shift

import (
	"testing"
	"time"

	"github.com/mustafaturan/shift/restrictors"
	"github.com/mustafaturan/shift/timers"
	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreaker(t *testing.T) {
	t.Run("without options", func(t *testing.T) {
		cb, err := NewCircuitBreaker("test")
		assert.IsType(t, &CircuitBreaker{}, cb)
		assert.NoError(t, err)
		assert.Equal(t, "test", cb.name)
		assert.Equal(t, StateClose, cb.state)
		assert.Equal(t, int64(0), cb.failure)
		assert.Equal(t, int64(0), cb.success)
		assert.Equal(t, int64(3), cb.failureMinRequests)
		assert.Equal(t, float64(99.9), cb.failureThreshold)
		assert.Equal(t, int64(2), cb.successThreshold)
		assert.Equal(t, 5*time.Second, cb.invocationTimeout)
		assert.NotNil(t, cb.resetAt)
		assert.NotNil(t, cb.resetTimer)
		assert.Equal(t, 0, len(cb.restrictors))
		assert.Equal(t, 0, len(cb.onStateChangeHandlers))
		assert.Equal(t, 0, len(cb.onFailureHandlers))
		assert.Equal(t, 0, len(cb.onSuccessHandlers))
	})

	t.Run("with valid options", func(t *testing.T) {
		cb, err := NewCircuitBreaker(
			"test",
			WithInitialState(StateHalfOpen),
			WithFailureThreshold(float64(99.99), 1),
		)
		assert.IsType(t, &CircuitBreaker{}, cb)
		assert.NoError(t, err)
		assert.Equal(t, StateHalfOpen, cb.state)
		assert.Equal(t, float64(99.99), cb.failureThreshold)
	})

	t.Run("with invalid options", func(t *testing.T) {
		cb, err := NewCircuitBreaker(
			"test",
			WithSuccessThreshold(int64(-1)),
		)
		assert.Nil(t, cb)
		assert.Error(t, err)
	})
}

func TestWithInitialState(t *testing.T) {
	cb, _ := NewCircuitBreaker("test")

	opt := WithInitialState(StateOpen)
	err := opt(cb)
	assert.NoError(t, err)
	assert.Equal(t, StateOpen, cb.state)
}

func TestWithFailureThreshold(t *testing.T) {
	t.Run("valid option values", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		opt := WithFailureThreshold(float64(99.99), 5)
		err := opt(cb)
		assert.NoError(t, err)
		assert.Equal(t, float64(99.99), cb.failureThreshold)
		assert.Equal(t, int64(5), cb.failureMinRequests)
	})

	t.Run("invalid option value for threshold", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		opt := WithFailureThreshold(float64(-0.1), 5)
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.NotEqual(t, float64(-0.1), cb.failureThreshold)
		assert.NotEqual(t, int64(5), cb.failureMinRequests)
	})
	t.Run("invalid option value for min requests", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		opt := WithFailureThreshold(float64(99.99), -1)
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.NotEqual(t, float64(99.99), cb.failureThreshold)
		assert.NotEqual(t, int64(-1), cb.failureMinRequests)
	})
}

func TestWithSuccessThreshold(t *testing.T) {
	t.Run("valid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		opt := WithSuccessThreshold(int64(9))
		err := opt(cb)
		assert.NoError(t, err)
		assert.Equal(t, int64(9), cb.successThreshold)
	})

	t.Run("invalid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		opt := WithSuccessThreshold(int64(-3))
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.NotEqual(t, int64(-3), cb.successThreshold)
	})
}

func TestWithTimeoutDuration(t *testing.T) {
	cb, _ := NewCircuitBreaker("test")

	duration := 5 * time.Second
	opt := WithInvocationTimeout(duration)
	err := opt(cb)
	assert.NoError(t, err)
	assert.Equal(t, duration, cb.invocationTimeout)
}

func TestWithResetTimer(t *testing.T) {
	cb, _ := NewCircuitBreaker("test")

	timer := timers.NewConstantTimer(time.Duration(5))
	opt := WithResetTimer(timer)
	err := opt(cb)
	assert.NoError(t, err)
	assert.Equal(t, timer, cb.resetTimer)
}

func TestWithRestrictors(t *testing.T) {
	t.Run("valid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		restrictor1, _ := restrictors.NewConcurrentRunRestrictor("test", int64(3))
		restrictor2, _ := restrictors.NewConcurrentRunRestrictor("test", int64(5))
		opt := WithRestrictors(restrictor1, restrictor2)
		err := opt(cb)
		assert.NoError(t, err)
		assert.Equal(t, []Restrictor{restrictor1, restrictor2}, cb.restrictors)
	})

	t.Run("invalid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var restrictor Restrictor
		opt := WithRestrictors(restrictor)
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Equal(t, []Restrictor{}, cb.restrictors)
	})
}

func TestWithOnStateChangeHandlers(t *testing.T) {
	t.Run("valid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var handler1 OnStateChange = func(_, _ State) {}
		var handler2 OnStateChange = func(_, _ State) {}
		var handler3 OnStateChange = func(_, _ State) {}
		opt := WithOnStateChangeHandlers(handler1, handler2, handler3)
		err := opt(cb)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(cb.onStateChangeHandlers))
	})

	t.Run("invalid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var handler OnStateChangeHandler
		opt := WithOnStateChangeHandlers(handler)
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Equal(t, []OnStateChangeHandler{}, cb.onStateChangeHandlers)
	})
}

func TestWithOnFailureHandlers(t *testing.T) {
	t.Run("valid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var handler1 OnFailure = func(_ State, err error) {}
		var handler2 OnFailure = func(_ State, err error) {}
		var handler3 OnFailure = func(_ State, err error) {}
		opt := WithOnFailureHandlers(handler1, handler2, handler3)
		err := opt(cb)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(cb.onFailureHandlers))
	})

	t.Run("invalid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var handler OnFailureHandler
		opt := WithOnFailureHandlers(handler)
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Equal(t, []OnStateChangeHandler{}, cb.onStateChangeHandlers)
	})
}

func TestWithOnSuccessHandlers(t *testing.T) {
	t.Run("valid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var handler1 OnSuccess = func(_ interface{}) {}
		var handler2 OnSuccess = func(_ interface{}) {}
		var handler3 OnSuccess = func(_ interface{}) {}
		opt := WithOnSuccessHandlers(handler1, handler2, handler3)
		err := opt(cb)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(cb.onSuccessHandlers))
	})

	t.Run("invalid option value", func(t *testing.T) {
		cb, _ := NewCircuitBreaker("test")
		var handler OnSuccessHandler
		opt := WithOnSuccessHandlers(handler)
		err := opt(cb)
		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Equal(t, []OnStateChangeHandler{}, cb.onStateChangeHandlers)
	})
}

func TestOnStateChange(t *testing.T) {
	var fn OnStateChange
	assert.Panics(t, func() { fn.Handle(StateOpen, StateHalfOpen) })
}

func TestOnFailure(t *testing.T) {
	var fn OnFailure
	assert.Panics(t, func() { fn.Handle(StateHalfOpen, nil) })
}

func TestOnSuccess(t *testing.T) {
	var fn OnSuccess
	assert.Panics(t, func() { fn.Handle(nil) })
}

func TestInvalidOptionError(t *testing.T) {
	err := &InvalidOptionError{Name: "test", Type: "any"}
	assert.EqualError(t, err, "invalid option provided for test, must be any")
}
