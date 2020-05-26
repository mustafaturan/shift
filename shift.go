// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"context"
	"sync"
	"time"

	"github.com/mustafaturan/shift/counter"
	"github.com/mustafaturan/shift/timer"
)

const (
	// Version matches with the current version of the package
	Version = "1.0.0-beta"
)

// Shift is an optioned circuit breaker implementation
type Shift struct {
	mutex sync.RWMutex

	// Name is an identity for the circuit breaker to increase observability
	// on failures
	name string

	// State of the circuit breaker. It can have open, half-open, close values
	state State

	// Counter is behaviour for circuit breaker metrics which supports the basic
	// increment and reset operations
	counter Counter

	// ResetTimer is a duration builder for resetting the state
	resetTimer Timer

	// Resetter holds the timer which resets the circuit breaker state
	resetter *time.Timer

	// Invokers holds invokers per state. Invokers are also
	invokers map[State]invoker

	// Trippers
	halfOpenCloser SuccessHandler
	halfOpenOpener FailureHandler
	closeOpener    FailureHandler

	successHandlers map[State][]SuccessHandler
	failureHandlers map[State][]FailureHandler

	// Restrictors are pre-callback actions which applies right before the
	// invocations. The restrictors can block the invocation with error returns.
	restrictors []Restrictor

	// StateChangeHandlers are callbacks which called on every state changes
	stateChangeHandlers []StateChangeHandler
}

const (
	// optionDefaultInitialState default initial state
	optionDefaultInitialState = StateClose

	// optionDefaultResetTimer default wait time
	optionDefaultResetTimer = 15 * time.Second

	// optionDefaultInvocationTimeout default invocation timeout duration
	optionDefaultInvocationTimeout = 5 * time.Second

	// optionDefaultCounterCapacity default capacity for counter
	optionDefaultCounterCapacity = 10

	// optionDefaultCounterBucketDuration default duration for counter buckets
	optionDefaultCounterBucketDuration = time.Second

	// optionDefaultMinSuccessRatioForCloseOpener minimum success ratio required
	// to keep the state as is
	optionDefaultMinSuccessRatioForCloseOpener = 90.0

	// optionDefaultMinSuccessRatioForHalfOpenOpener minimum success ratio required
	// to keep the state as is
	optionDefaultMinSuccessRatioForHalfOpenOpener = 70.0

	// optionDefaultMinSuccessRatioForHalfOpenCloser minimum success ratio required
	// to trip to 'close' state
	optionDefaultMinSuccessRatioForHalfOpenCloser = 85.0

	// optionDefaultMinRequests
	optionDefaultMinRequests = 10
)

// New inits a new Circuit Breaker with given name and options
func New(name string, opts ...Option) (*Shift, error) {
	s := &Shift{
		name:     name,
		state:    optionDefaultInitialState,
		resetter: time.AfterFunc(time.Microsecond, func() {}),
		invokers: map[State]invoker{
			StateClose: &onCloseInvoker{
				timeout: optionDefaultInvocationTimeout,
			},
			StateHalfOpen: &onHalfOpenInvoker{
				timeout: optionDefaultInvocationTimeout,
			},
			StateOpen: &onOpenInvoker{},
		},
		failureHandlers: map[State][]FailureHandler{
			StateClose:    make([]FailureHandler, 0),
			StateHalfOpen: make([]FailureHandler, 0),
			StateOpen:     make([]FailureHandler, 0),
		},
		successHandlers: map[State][]SuccessHandler{
			StateClose:    make([]SuccessHandler, 0),
			StateHalfOpen: make([]SuccessHandler, 0),
			StateOpen:     make([]SuccessHandler, 0),
		},
		stateChangeHandlers: make([]StateChangeHandler, 0),
		restrictors:         make([]Restrictor, 0),
	}

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			return nil, err
		}
	}

	// Init the default counter if not specified
	if s.counter == nil {
		s.counter, _ = counter.NewTimeBucketCounter(
			optionDefaultCounterCapacity,
			optionDefaultCounterBucketDuration,
		)
	}

	if s.resetTimer == nil {
		s.resetTimer, _ = timer.NewConstantTimer(optionDefaultResetTimer)
	}

	s.invokers[StateClose].(*onCloseInvoker).timeoutCallback = func() {
		s.counter.Increment(metricTimeout)
	}
	s.invokers[StateHalfOpen].(*onHalfOpenInvoker).timeoutCallback = func() {
		s.counter.Increment(metricTimeout)
	}
	s.invokers[StateOpen].(*onOpenInvoker).rejectCallback = func() {
		s.counter.Increment(metricReject)
	}

	if s.closeOpener == nil {
		_ = WithOpener(StateClose, optionDefaultMinSuccessRatioForCloseOpener, optionDefaultMinRequests)(s)
	}
	s.failureHandlers[StateClose] = append([]FailureHandler{s.closeOpener}, s.failureHandlers[StateClose]...)

	if s.halfOpenOpener == nil {
		_ = WithOpener(StateHalfOpen, optionDefaultMinSuccessRatioForHalfOpenOpener, optionDefaultMinRequests)(s)
	}
	s.failureHandlers[StateHalfOpen] = append([]FailureHandler{s.closeOpener}, s.failureHandlers[StateHalfOpen]...)

	if s.halfOpenCloser == nil {
		_ = WithCloser(optionDefaultMinSuccessRatioForHalfOpenCloser, optionDefaultMinRequests)(s)
	}
	s.successHandlers[StateHalfOpen] = append([]SuccessHandler{s.halfOpenCloser}, s.successHandlers[StateHalfOpen]...)

	return s, nil
}

// WithInitialState builds option to set initial state
func WithInitialState(state State) Option {
	return func(s *Shift) error {
		s.state = state
		return nil
	}
}

// WithInvocationTimeout builds option to set invocation timeout duration
func WithInvocationTimeout(duration time.Duration) Option {
	return func(s *Shift) error {
		s.invokers[StateClose].(*onCloseInvoker).timeout = duration
		s.invokers[StateHalfOpen].(*onHalfOpenInvoker).timeout = duration
		return nil
	}
}

// WithResetTimer builds option to set reset timer
func WithResetTimer(t Timer) Option {
	return func(s *Shift) error {
		s.resetTimer = t
		return nil
	}
}

// WithCounter builds option to set stats counter
func WithCounter(c Counter) Option {
	return func(s *Shift) error {
		s.counter = c
		return nil
	}
}

// WithRestrictors builds option to set restrictors to restrict the invocations
// Restrictors does not effect the current state, but they can block the
// invocation depending on its own internal state values. If a restrictor blocks
// an invocation then it returns an error and `On Failure Handlers` get executed
// in order.
func WithRestrictors(restrictors ...Restrictor) Option {
	return func(s *Shift) error {
		for _, r := range restrictors {
			if r == nil {
				return &InvalidOptionError{
					Name:    "restrictor",
					Message: "can't be nil",
				}
			}
		}
		s.restrictors = restrictors
		return nil
	}
}

// WithStateChangeHandlers builds option to set state change handlers, the
// provided handlers will be evaluate in the given order as option
func WithStateChangeHandlers(handlers ...StateChangeHandler) Option {
	return func(s *Shift) error {
		for _, h := range handlers {
			if h == nil {
				return &InvalidOptionError{
					Name:    "on state change handler",
					Message: "can't be nil",
				}
			}
		}
		s.stateChangeHandlers = handlers
		return nil
	}
}

// WithSuccessHandlers builds option to set on failure handlers, the provided
// handlers will be evaluate in the given order as option
func WithSuccessHandlers(state State, handlers ...SuccessHandler) Option {
	return func(s *Shift) error {
		for _, h := range handlers {
			if h == nil {
				return &InvalidOptionError{
					Name:    "on success handler",
					Message: "can't be nil",
				}
			}
		}

		s.successHandlers[state] = append(s.successHandlers[state], handlers...)
		return nil
	}
}

// WithFailureHandlers builds option to set on failure handlers, the
// provided handlers will be evaluate in the given order as option
func WithFailureHandlers(state State, handlers ...FailureHandler) Option {
	return func(s *Shift) error {
		for _, h := range handlers {
			if h == nil {
				return &InvalidOptionError{
					Name:    "failure handler",
					Message: "can't be nil",
				}
			}
		}
		s.failureHandlers[state] = append(s.failureHandlers[state], handlers...)
		return nil
	}
}

// WithOpener builds an option to set the default failure criteria to trip to
// 'open' state. (If the failure criteria matches then the circuit breaker
// trips to the 'open' state.)
//
// As runtime behaviour, it prepends a failure handler for the given state to
// trip circuit breaker into the 'open' state when the given thresholds reached.
//
// Definitions of the params are
// state: StateClose, StateHalfOpen
// minSuccessRatio: min success ratio ratio to keep the Circuit Breaker as is
// minRequests: min number of requests before checking the ratio
//
// Params with example:
// state: StateClose, minSuccessRatio: 95%, minRequests: 10
// The above configuration means that:
// On 'close' state, at min 10 requests, if it calculates the success ratio less
// than or equal to 95% then will trip to 'open' state
func WithOpener(state State, minSuccessRatio float32, minRequests uint32) Option {
	return func(s *Shift) error {
		if !state.isClose() && !state.isHalfOpen() {
			return &InvalidOptionError{
				Name:    "state for failure criteria",
				Message: "can only be applied to 'close' and 'half open' states",
			}
		}

		if minSuccessRatio <= 0.0 || minSuccessRatio > 100.0 {
			return &InvalidOptionError{
				Name:    "min success ratio to trip to 'open' state",
				Message: "can be greater than 0.0 and less than equal to 100.0",
			}
		}

		if minRequests < 1 {
			return &InvalidOptionError{
				Name:    "min requests to check success ratio",
				Message: "must be positive int",
			}
		}

		var handler OnFailure = func(ctx context.Context, _ error) {
			stats := ctx.Value(CtxStats).(Stats)
			requests := stats.SuccessCount + stats.FailureCount - stats.RejectCount
			if requests < minRequests {
				return
			}

			ratio := float32(stats.SuccessCount) / float32(requests) * 100

			if ratio < minSuccessRatio {
				_ = s.Trip(StateOpen, &FailureThresholdReachedError{})
			}
		}

		if state.isHalfOpen() {
			s.halfOpenOpener = handler
		} else {
			s.closeOpener = handler
		}

		return nil
	}
}

// WithCloser builds an option to set the default success criteria trip to
// 'close' state. (If the success criteria matches then the circuit breaker
// trips to the 'close' state.)
//
// As runtime behaviour, it appends a success handler for the given state to
// trip circuit breaker into the 'close' state when the given thresholds reached
//
// Definitions of the params are
// state: StateHalfOpen(always half-open it is a hidden param)
// minSuccessRatio: min success ratio to trip the circuit breaker to close state
// minRequests: min number of requests before checking the ratio
//
// Params with example:
// state: StateHalfOpen, minSuccessRatio: 99.5%, minRequests: 1000
// The above configuration means that:
// On 'half-open' state, at min 1000 requests, if it counts 995 success then
// will trip to 'close' state
func WithCloser(minSuccessRatio float32, minRequests uint32) Option {
	return func(s *Shift) error {
		if minSuccessRatio <= 0.0 || minSuccessRatio > 100.0 {
			return &InvalidOptionError{
				Name:    "min success ratio to trip to 'close' state",
				Message: "can be greater than 0.0 and less than equal to 100.0",
			}
		}

		if minRequests < 1 {
			return &InvalidOptionError{
				Name:    "min requests to check success ratio",
				Message: "must be positive int",
			}
		}

		var handler OnSuccess = func(ctx context.Context, _ interface{}) {
			stats := ctx.Value(CtxStats).(Stats)
			requests := stats.SuccessCount + stats.FailureCount - stats.RejectCount
			if requests < minRequests {
				return
			}

			ratio := float32(stats.SuccessCount) / float32(requests) * 100
			if ratio >= minSuccessRatio {
				_ = s.Trip(StateClose)
			}
		}

		s.halfOpenCloser = handler

		return nil
	}
}
