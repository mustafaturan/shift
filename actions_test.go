package shift

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/mustafaturan/shift/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("with restrictor rejection", func(t *testing.T) {
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		restrictor := mock.NewMockRestrictor(ctrl)

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateClose),
			WithRestrictors(restrictor),
		)

		require.NoError(t, err)

		restrictor.
			EXPECT().
			Check(gomock.Any()).
			Return(false, errors.New("too many reqs"))

		restrictor.
			EXPECT().
			Defer()

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(map[string]uint32{"success": 0, "failure": 1, "rejects": 1})

		counter.
			EXPECT().
			Increment(metricReject)

		counter.
			EXPECT().
			Increment(metricFailure)

		ctx := context.Background()
		var o Operate = func(context.Context) (interface{}, error) {
			return nil, nil
		}

		res, err := s.Run(ctx, o)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("reject on open state", func(t *testing.T) {
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		restrictor := mock.NewMockRestrictor(ctrl)

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateOpen),
			WithRestrictors(restrictor),
		)

		require.NoError(t, err)

		restrictor.
			EXPECT().
			Check(gomock.Any()).
			Return(true, nil)

		restrictor.
			EXPECT().
			Defer()

		counter.
			EXPECT().
			Increment(metricReject)

		counter.
			EXPECT().
			Increment(metricFailure)

		ctx := context.Background()
		var o Operate = func(context.Context) (interface{}, error) {
			return nil, nil
		}

		res, err := s.Run(ctx, o)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("invocation successes on operate fn", func(t *testing.T) {
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		restrictor := mock.NewMockRestrictor(ctrl)
		var called bool
		var handler OnSuccess = func(ctx context.Context, res interface{}) {
			assert.IsType(t, Stats{}, ctx.Value(CtxStats))
			assert.Equal(t, StateClose, ctx.Value(CtxState))

			assert.Equal(t, "welldone1", res.(string))
			called = true
		}

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateClose),
			WithRestrictors(restrictor),
			WithSuccessHandlers(StateClose, handler),
		)

		require.NoError(t, err)

		restrictor.
			EXPECT().
			Check(gomock.Any()).
			Return(true, nil)

		restrictor.
			EXPECT().
			Defer()

		counter.
			EXPECT().
			Increment(metricSuccess)

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(map[string]uint32{})

		ctx := context.Background()
		var o Operate = func(context.Context) (interface{}, error) {
			return "welldone1", nil
		}

		res, err := s.Run(ctx, o)
		assert.NoError(t, err)
		assert.Equal(t, "welldone1", res.(string))
		assert.Equal(t, true, called)
	})

	t.Run("invocation successes on operate fn without success callbacks", func(t *testing.T) {
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateClose),
		)

		require.NoError(t, err)

		counter.
			EXPECT().
			Increment(metricSuccess)

		ctx := context.Background()
		var o Operate = func(context.Context) (interface{}, error) {
			return "welldone2", nil
		}

		res, err := s.Run(ctx, o)
		assert.NoError(t, err)
		assert.Equal(t, "welldone2", res.(string))
	})

	t.Run("invocation fails on operate fn", func(t *testing.T) {
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		restrictor := mock.NewMockRestrictor(ctrl)
		failureErr := errors.New("failed")
		var called bool
		var handler OnFailure = func(ctx context.Context, err error) {
			assert.IsType(t, Stats{}, ctx.Value(CtxStats))
			assert.Equal(t, StateClose, ctx.Value(CtxState))

			assert.EqualError(t, err, "circuit breaker(test) invocation failed with failed")
			called = true
		}

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateClose),
			WithRestrictors(restrictor),
			WithFailureHandlers(StateClose, handler),
		)

		require.NoError(t, err)

		restrictor.
			EXPECT().
			Check(gomock.Any()).
			Return(true, nil)

		restrictor.
			EXPECT().
			Defer()

		counter.
			EXPECT().
			Increment(metricFailure)

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(map[string]uint32{})

		ctx := context.Background()
		var o Operate = func(context.Context) (interface{}, error) {
			return nil, failureErr
		}

		res, err := s.Run(ctx, o)
		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, true, called)
	})

	t.Run("invocation fails with timeout", func(t *testing.T) {
		// same reaction on both close and half-open states
		states := []State{StateClose, StateHalfOpen}
		for _, currentState := range states {
			state := currentState
			timer := mock.NewMockTimer(ctrl)
			counter := mock.NewMockCounter(ctrl)
			restrictor := mock.NewMockRestrictor(ctrl)
			var called bool
			var handler OnFailure = func(ctx context.Context, err error) {
				assert.IsType(t, Stats{}, ctx.Value(CtxStats))
				assert.Equal(t, state, ctx.Value(CtxState))

				assert.EqualError(t, err, "circuit breaker(test) invocation failed with invocation timeout on 100ms")
				called = true
			}

			s, err := New(
				name,
				WithCounter(counter),
				WithResetTimer(timer),
				WithInitialState(state),
				WithRestrictors(restrictor),
				WithFailureHandlers(state, handler),
				WithInvocationTimeout(100*time.Millisecond),
			)

			require.NoError(t, err)

			restrictor.
				EXPECT().
				Check(gomock.Any()).
				Return(true, nil)

			restrictor.
				EXPECT().
				Defer()

			counter.
				EXPECT().
				Increment(metricTimeout)

			counter.
				EXPECT().
				Increment(metricFailure)

			counter.
				EXPECT().
				Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
				Return(map[string]uint32{})

			ctx := context.Background()
			var o Operate = func(context.Context) (interface{}, error) {
				time.Sleep(110 * time.Millisecond)
				return "welldone3", nil
			}

			res, err := s.Run(ctx, o)
			assert.Error(t, err)
			assert.Nil(t, res)
			assert.Equal(t, true, called)
		}
	})
}

func TestTrip(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stats := map[string]uint32{}

	t.Run("to the current state", func(t *testing.T) {
		var called bool
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		var stateChangeHandler OnStateChange = func(from, to State, stats Stats) {
			called = true
		}

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateClose),
			WithStateChangeHandlers(stateChangeHandler),
		)
		require.NoError(t, err)

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(stats)

		err = s.Trip(StateClose)

		assert.Error(t, err)
		assert.Equal(t, false, called)
	})

	t.Run("to close state", func(t *testing.T) {
		var called bool
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		var stateChangeHandler OnStateChange = func(from, to State, stats Stats) {
			assert.Equal(t, StateHalfOpen, from)
			assert.Equal(t, StateClose, to)
			called = true
		}

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateHalfOpen),
			WithStateChangeHandlers(stateChangeHandler),
		)
		require.NoError(t, err)

		timer.
			EXPECT().
			Reset()

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(stats)

		counter.
			EXPECT().
			Reset()

		err = s.Trip(StateClose)

		assert.NoError(t, err)
		assert.Equal(t, StateClose, s.currentState())
		assert.Equal(t, true, called)
	})

	t.Run("to half-open state", func(t *testing.T) {
		var called bool
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		var stateChangeHandler OnStateChange = func(from, to State, stats Stats) {
			assert.Equal(t, StateOpen, from)
			assert.Equal(t, StateHalfOpen, to)
			called = true
		}

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateOpen),
			WithStateChangeHandlers(stateChangeHandler),
		)
		require.NoError(t, err)

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(stats)

		counter.
			EXPECT().
			Reset()

		err = s.Trip(StateHalfOpen)

		assert.NoError(t, err)
		assert.Equal(t, StateHalfOpen, s.currentState())
		assert.Equal(t, true, called)
	})

	t.Run("to open state", func(t *testing.T) {
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateHalfOpen),
		)
		require.NoError(t, err)

		reason := errors.New("reason")

		timer.
			EXPECT().
			Next(reason).
			Return(time.Second)

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(stats).
			Times(2)

		counter.
			EXPECT().
			Reset().
			Times(2)

		err = s.Trip(StateOpen, reason)

		assert.NoError(t, err)
		assert.Equal(t, StateOpen, s.currentState())

		// Trips to half-open state after 1.0+ seconds
		time.Sleep(1100 * time.Millisecond)

		assert.NoError(t, err)
		assert.Equal(t, StateHalfOpen, s.currentState())
	})

	t.Run("to unknown state", func(t *testing.T) {
		var called bool
		timer := mock.NewMockTimer(ctrl)
		counter := mock.NewMockCounter(ctrl)
		var stateChangeHandler OnStateChange = func(from, to State, stats Stats) {
			called = true
		}

		s, err := New(
			name,
			WithCounter(counter),
			WithResetTimer(timer),
			WithInitialState(StateOpen),
			WithStateChangeHandlers(stateChangeHandler),
		)
		require.NoError(t, err)

		counter.
			EXPECT().
			Stats(metricSuccess, metricFailure, metricTimeout, metricReject).
			Return(stats)

		err = s.Trip(StateUnknown)
		assert.Error(t, err)
		assert.Equal(t, false, called)
	})
}
