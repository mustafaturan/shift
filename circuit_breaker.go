// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"fmt"
	"sync"
	"time"

	"github.com/mustafaturan/shift/timers"
)

// CircuitBreaker is circuit breaker implementation
type CircuitBreaker struct {
	mutex sync.RWMutex

	name  string
	state State

	failure int64
	success int64
	resetAt time.Time

	failureMinRequests    int64
	failureRatioThreshold float64
	successThreshold      int64
	invocationTimeout     time.Duration
	resetTimer            Timer

	restrictors []Restrictor

	onStateChangeHandlers []OnStateChangeHandler
	onFailureHandlers     []OnFailureHandler
	onSuccessHandlers     []OnSuccessHandler
}

// invocation is a type for holding invocation result
type invocation struct {
	res interface{}
	err error
}

// Option is a type for circuit breaker options
type Option func(*CircuitBreaker) error

const (
	// optionDefaultInitialState default initial state
	optionDefaultInitialState = StateClose

	// optionDefaultFailureMinRequests default failure min requests
	optionDefaultFailureMinRequests = int64(3)

	// optionDefaultFailureThreshold default failure threshold
	optionDefaultFailureThreshold = float64(99.9)

	// optionDefaultSuccessThreshold default success threshold
	optionDefaultSuccessThreshold = int64(2)

	// optionDefaultResetTimer default wait time
	optionDefaultResetTimer = 3 * time.Second

	// optionDefaultInvocationTimeout default invocation timeout duration
	optionDefaultInvocationTimeout = 5 * time.Second
)

// NewCircuitBreaker inits a new CircuitBreaker with given name and options
func NewCircuitBreaker(name string, opts ...Option) (*CircuitBreaker, error) {
	cb := &CircuitBreaker{
		name:                  name,
		state:                 optionDefaultInitialState,
		failureMinRequests:    optionDefaultFailureMinRequests,
		failureRatioThreshold: optionDefaultFailureThreshold,
		successThreshold:      optionDefaultSuccessThreshold,
		invocationTimeout:     optionDefaultInvocationTimeout,
		resetTimer:            timers.NewConstantTimer(optionDefaultResetTimer),
		restrictors:           []Restrictor{},
		onStateChangeHandlers: []OnStateChangeHandler{},
		onFailureHandlers:     []OnFailureHandler{},
		onSuccessHandlers:     []OnSuccessHandler{},
	}

	for _, opt := range opts {
		err := opt(cb)
		if err != nil {
			return nil, err
		}
	}

	return cb, nil
}

// WithInitialState builds option to set initial state
func WithInitialState(s State) Option {
	return func(cb *CircuitBreaker) error {
		cb.state = s
		return nil
	}
}

// WithFailureThreshold builds option to set threshold value as percentage for
// successes over all requests
func WithFailureThreshold(threshold float64, minRequests int64) Option {
	return func(cb *CircuitBreaker) error {
		if threshold < 1 {
			return &InvalidOptionError{
				Name: "failure threshold success rate",
				Type: "positive float 32",
			}
		}
		if minRequests < 1 {
			return &InvalidOptionError{
				Name: "minimum requests threshold",
				Type: "positive integer",
			}
		}
		cb.failureRatioThreshold = threshold
		cb.failureMinRequests = minRequests
		return nil
	}
}

// WithSuccessThreshold builds option to set threshold value for success
func WithSuccessThreshold(threshold int64) Option {
	return func(cb *CircuitBreaker) error {
		if threshold < 1 {
			return &InvalidOptionError{
				Name: "success threshold",
				Type: "positive integer",
			}
		}
		cb.successThreshold = threshold
		return nil
	}
}

// WithInvocationTimeout builds option to set invocation timeout duration
func WithInvocationTimeout(duration time.Duration) Option {
	return func(cb *CircuitBreaker) error {
		cb.invocationTimeout = duration
		return nil
	}
}

// WithResetTimer builds option to set reset time on close state
func WithResetTimer(t Timer) Option {
	return func(cb *CircuitBreaker) error {
		cb.resetTimer = t
		return nil
	}
}

// WithRestrictors builds option to set restrictors to restrict the invocations
// Restrictors does not effect the current state, but they can block the
// invocation depending on its own internal state values. If a restrictor blocks
// an invocation then it returns an error and `On Failure Handlers` get executed
// in order.
func WithRestrictors(restrictors ...Restrictor) Option {
	return func(cb *CircuitBreaker) error {
		for _, r := range restrictors {
			if r == nil {
				return &InvalidOptionError{
					Name: "restrictor",
					Type: "can't be nil",
				}
			}
		}
		cb.restrictors = restrictors
		return nil
	}
}

// WithOnStateChangeHandlers builds option to set state change handlers, the
// provided handlers will be evaluate in the given order as option
func WithOnStateChangeHandlers(handlers ...OnStateChangeHandler) Option {
	return func(cb *CircuitBreaker) error {
		for _, h := range handlers {
			if h == nil {
				return &InvalidOptionError{
					Name: "on state change handler",
					Type: "can't be nil",
				}
			}
		}
		cb.onStateChangeHandlers = handlers
		return nil
	}
}

// WithOnFailureHandlers builds option to set on failure handlers, the
// provided handlers will be evaluate in the given order as option
func WithOnFailureHandlers(handlers ...OnFailureHandler) Option {
	return func(cb *CircuitBreaker) error {
		for _, h := range handlers {
			if h == nil {
				return &InvalidOptionError{
					Name: "on failure handler",
					Type: "can't be nil",
				}
			}
		}
		cb.onFailureHandlers = handlers
		return nil
	}
}

// WithOnSuccessHandlers builds option to set on failure handlers, the
// provided handlers will be evaluate in the given order as option
func WithOnSuccessHandlers(handlers ...OnSuccessHandler) Option {
	return func(cb *CircuitBreaker) error {
		for _, h := range handlers {
			if h == nil {
				return &InvalidOptionError{
					Name: "on success handler",
					Type: "can't be nil",
				}
			}
		}
		cb.onSuccessHandlers = handlers
		return nil
	}
}

// Timer is an interface to set reset time duration dynamically depending on
// the occurred error on the invocation
type Timer interface {
	// Next returns the current duration and sets the next duration according to
	// the given error
	Next(error) time.Duration

	// Reset resets the current duration to the initial duration
	Reset()
}

// Restrictor allows adding restriction to circuit breaker
type Restrictor interface {
	// Check checks if restriction allows to run current invocation and errors if
	// not allowed the invocation
	Check() (bool, error)

	// Defer executes exit rules of the restrictor right after the run process
	Defer()
}

// OnStateChangeHandler is an interface to handle state change events
type OnStateChangeHandler interface {
	Handle(from, to State)
}

// OnStateChange is a function to run on any state change invocation
type OnStateChange func(from, to State)

// Handle implements OnStateChangeHandler for OnStateChange func
func (fn OnStateChange) Handle(from, to State) {
	fn(from, to)
}

// OnFailureHandler is an interface to handle failure events
type OnFailureHandler interface {
	Handle(State, error)
}

// OnFailure is a function to run on any error like timeout and invocation errors
type OnFailure func(State, error)

// Handle implements OnFailureHandler for OnFailure func
func (fn OnFailure) Handle(s State, err error) {
	fn(s, err)
}

// OnSuccessHandler is an interface to handle success events
type OnSuccessHandler interface {
	Handle(interface{})
}

// OnSuccess is a function to run on any successful invocation
type OnSuccess func(interface{})

// Handle implements OnSuccessHandler for OnSuccess func
func (fn OnSuccess) Handle(data interface{}) {
	fn(data)
}

// InvalidOptionError is a error tyoe for options
type InvalidOptionError struct {
	Name string
	Type string
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf(
		"invalid option provided for %s, must be %s",
		e.Name,
		e.Type,
	)
}
