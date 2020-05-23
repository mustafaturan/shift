package shift

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mustafaturan/shift/mock"
	"github.com/mustafaturan/shift/restrictor"
	"github.com/stretchr/testify/assert"
)

const (
	name = "test"
)

func TestVersion(t *testing.T) {
	assert.Equal(t, Version, "1.0.0-alpha")
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("without optional reset timer", func(t *testing.T) {
		s, err := New(name)
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.resetTimer)
	})

	t.Run("without optional counter", func(t *testing.T) {
		s, err := New(name)
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.NotNil(t, s.counter)
	})

	t.Run("with required options", func(t *testing.T) {
		s, err := New(name, WithCounter(counter), WithResetTimer(timer))
		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.IsType(t, &Shift{}, s)

		t.Run("name", func(t *testing.T) {
			assert.Equal(t, name, s.name)
		})

		t.Run("initial state", func(t *testing.T) {
			assert.Equal(t, StateClose, s.state)
		})

		t.Run("initial reset at", func(t *testing.T) {
			assert.NotNil(t, s.resetter)
			assert.False(t, s.resetter.Stop())
		})

		t.Run("invokers", func(t *testing.T) {
			t.Run("close state", func(t *testing.T) {
				invoker, ok := s.invokers[StateClose].(*onCloseInvoker)

				assert.True(t, ok)
				assert.NotNil(t, invoker)
				assert.Equal(t, optionDefaultInvocationTimeout, invoker.timeout)
				assert.NotNil(t, invoker.timeoutCallback)
			})

			t.Run("half-open state", func(t *testing.T) {
				invoker, ok := s.invokers[StateHalfOpen].(*onHalfOpenInvoker)

				assert.True(t, ok)
				assert.NotNil(t, invoker)
				assert.Equal(t, optionDefaultInvocationTimeout, invoker.timeout)
				assert.NotNil(t, invoker.timeoutCallback)
			})

			t.Run("open state", func(t *testing.T) {
				invoker, ok := s.invokers[StateOpen].(*onOpenInvoker)

				assert.True(t, ok)
				assert.NotNil(t, invoker)
				assert.NotNil(t, invoker.rejectCallback)
			})
		})

		t.Run("restrictors", func(t *testing.T) {
			assert.Equal(t, 0, len(s.restrictors))
		})

		t.Run("state change handlers", func(t *testing.T) {
			assert.Equal(t, 0, len(s.stateChangeHandlers))
		})
	})
}

func TestWithInitialState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)
	state := StateOpen

	s, err := New(
		name,
		WithCounter(counter),
		WithResetTimer(timer),
		WithInitialState(state),
	)
	assert.NoError(t, err)
	assert.Equal(t, state, s.state)
}

func TestWithInvokationTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)
	duration := 90 * time.Second
	s, err := New(
		name,
		WithCounter(counter),
		WithResetTimer(timer),
		WithInvocationTimeout(duration),
	)

	assert.NoError(t, err)
	assert.Equal(t, duration, s.invokers[StateClose].(*deadlineInvoker).timeout)
	assert.Equal(t, duration, s.invokers[StateHalfOpen].(*deadlineInvoker).timeout)
}

func TestWithRestrictors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("with a nil restrictor", func(t *testing.T) {
		var restrictor Restrictor
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithRestrictors(restrictor),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with valid options", func(t *testing.T) {
		restrictor := &restrictor.ConcurrentRunRestrictor{}
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithRestrictors(restrictor),
		)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(s.restrictors))
		assert.Equal(t, restrictor, s.restrictors[0])
	})
}

func TestWithOnStateChangeHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("with a nil state change handler", func(t *testing.T) {
		var validHandler OnStateChange = func(_, _ State, _ Stats) {}
		var nilHandler StateChangeHandler
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithStateChangeHandlers(validHandler, nilHandler),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with valid options", func(t *testing.T) {
		var handler1 OnStateChange = func(_, _ State, _ Stats) {}
		var handler2 OnStateChange = func(_, _ State, _ Stats) {}
		var handler3 OnStateChange = func(_, _ State, _ Stats) {}
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithStateChangeHandlers(handler1, handler2, handler3),
		)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(s.stateChangeHandlers))
	})
}

func TestWithSuccessHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("with a nil state change handler", func(t *testing.T) {
		var validHandler OnSuccess = func(context.Context, interface{}) {}
		var nilHandler SuccessHandler
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithSuccessHandlers(StateClose, validHandler, nilHandler),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with valid options", func(t *testing.T) {
		var handler1 OnSuccess = func(context.Context, interface{}) {}
		var handler2 OnSuccess = func(context.Context, interface{}) {}
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithSuccessHandlers(StateClose, handler1, handler2),
			WithSuccessHandlers(StateHalfOpen, handler1),
		)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(s.successHandlers[StateClose]))
		// Additional one from trippers
		assert.Equal(t, 1+1, len(s.successHandlers[StateHalfOpen]))
		assert.Equal(t, 0, len(s.successHandlers[StateOpen]))
	})
}

func TestWithFailureHandlers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("with a nil state change handler", func(t *testing.T) {
		var validHandler OnFailure = func(context.Context, error) {}
		var nilHandler FailureHandler
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithFailureHandlers(StateClose, validHandler, nilHandler),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with valid options", func(t *testing.T) {
		var handler1 OnFailure = func(context.Context, error) {}
		var handler2 OnFailure = func(context.Context, error) {}
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithFailureHandlers(StateClose, handler1, handler2),
			WithFailureHandlers(StateHalfOpen, handler2),
			WithFailureHandlers(StateOpen, handler1),
		)

		assert.NoError(t, err)
		// Additional one from trippers
		assert.Equal(t, 2+1, len(s.failureHandlers[StateClose]))
		// Additional one from trippers
		assert.Equal(t, 1+1, len(s.failureHandlers[StateHalfOpen]))
		assert.Equal(t, 1, len(s.failureHandlers[StateOpen]))
	})
}

func TestWithOpener(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("with invalid state", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithOpener(StateOpen, -1.0, 100),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with invalid threshold", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithOpener(StateClose, -1.0, 100),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with invalid min requests", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithOpener(StateHalfOpen, 0.5, 0),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with valid options", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithOpener(StateClose, 98.5, 100),
		)

		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, 1, len(s.failureHandlers[StateClose]))

		handler := s.failureHandlers[StateClose][0]

		t.Run("execute without matching the min requests criteria", func(t *testing.T) {
			stats := Stats{SuccessCount: 95, FailureCount: 2, RejectCount: 1}
			ctx := context.WithValue(context.Background(), CtxStats, stats)
			handler.Handle(ctx, nil)

			assert.Equal(t, StateClose, s.currentState())
		})

		t.Run("execute without matching the success threshold criteria", func(t *testing.T) {
			stats := Stats{SuccessCount: 998, FailureCount: 2, RejectCount: 0}
			ctx := context.WithValue(context.Background(), CtxStats, stats)
			handler.Handle(ctx, nil)

			assert.Equal(t, StateClose, s.currentState())
		})

		t.Run("execute with matched criteria", func(t *testing.T) {
			stats := Stats{SuccessCount: 984, FailureCount: 16}
			ctx := context.WithValue(context.Background(), CtxStats, stats)
			counter.
				EXPECT().
				Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
				Return(map[string]uint32{"success": stats.SuccessCount, "failure": stats.FailureCount})

			counter.
				EXPECT().
				Reset()

			timer.
				EXPECT().
				Next(gomock.Any()).
				Return(60 * time.Second)

			// Trips to open state on matched criteria
			handler.Handle(ctx, nil)

			assert.Equal(t, StateOpen, s.currentState())
		})
	})
}

func TestWithCloser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	timer := mock.NewMockTimer(ctrl)
	counter := mock.NewMockCounter(ctrl)

	t.Run("with invalid threshold", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithCloser(-1.0, 100),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with invalid min requests", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithCloser(98.0, 0),
		)

		assert.Error(t, err)
		assert.IsType(t, &InvalidOptionError{}, err)
		assert.Nil(t, s)
	})

	t.Run("with valid options", func(t *testing.T) {
		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateHalfOpen),
			WithCloser(98.0, 100),
		)

		assert.NoError(t, err)
		assert.NotNil(t, s)
		assert.Equal(t, 1, len(s.successHandlers[StateHalfOpen]))

		handler := s.successHandlers[StateHalfOpen][0]

		t.Run("execute without matching the min requests criteria", func(t *testing.T) {
			stats := Stats{SuccessCount: 98, FailureCount: 2, RejectCount: 1}
			ctx := context.WithValue(context.Background(), CtxStats, stats)
			handler.Handle(ctx, nil)

			assert.Equal(t, StateHalfOpen, s.currentState())
		})

		t.Run("execute without matching the min success ratio criteria", func(t *testing.T) {
			stats := Stats{SuccessCount: 97, FailureCount: 3, RejectCount: 0}
			ctx := context.WithValue(context.Background(), CtxStats, stats)
			handler.Handle(ctx, nil)

			assert.Equal(t, StateHalfOpen, s.currentState())
		})

		t.Run("execute with matched criteria", func(t *testing.T) {
			stats := Stats{SuccessCount: 98, FailureCount: 2}
			ctx := context.WithValue(context.Background(), CtxStats, stats)
			counter.
				EXPECT().
				Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
				Return(map[string]uint32{"success": stats.SuccessCount, "failure": stats.FailureCount})

			counter.
				EXPECT().
				Reset()

			timer.
				EXPECT().
				Reset()

			// Trips to close state on success criteria
			handler.Handle(ctx, nil)

			assert.Equal(t, StateClose, s.currentState())
		})
	})
}

func TestNilOnStateChange(t *testing.T) {
	var fn OnStateChange
	assert.Panics(t, func() { fn.Handle(StateOpen, StateHalfOpen, Stats{}) })
}

func TestNilOnFailure(t *testing.T) {
	var fn OnFailure
	assert.Panics(t, func() { fn.Handle(context.Background(), nil) })
}

func TestNilOnSuccess(t *testing.T) {
	var fn OnSuccess
	assert.Panics(t, func() { fn.Handle(context.Background(), nil) })
}
