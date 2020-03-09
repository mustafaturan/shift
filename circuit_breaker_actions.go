// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// Run executes the given func with circuit breaker
func (cb *CircuitBreaker) Run(ctx context.Context, o Operator) (interface{}, error) {
	for _, r := range cb.restrictors {
		defer r.Defer()
		ok, err := r.Check()
		if !ok {
			return cb.runReject(err)
		}
	}

	return cb.run(ctx, o)
}

// Override manually overrides the current state with side effects
func (cb *CircuitBreaker) Override(to State) {
	switch to {
	case StateClose:
		cb.close()
	case StateHalfOpen:
		cb.halfOpen()
	case StateOpen:
		cb.open(&CircuitBreakerOverrideError{Name: cb.name})
	}
}

// State returns current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return cb.state
}

// Operator is an interface for circuit breaker operations
type Operator interface {
	Execute(context.Context) (interface{}, error)
}

// Operate is a function that runs the operation
type Operate func(context.Context) (interface{}, error)

// Execute implements Operator interface for any Operate fn
func (o Operate) Execute(ctx context.Context) (interface{}, error) {
	return o(ctx)
}

// CircuitBreakerOverrideError is a error type for open state
type CircuitBreakerOverrideError struct {
	Name string
}

func (e *CircuitBreakerOverrideError) Error() string {
	return fmt.Sprintf(
		"circuit breaker(%s) is overridden as open",
		e.Name,
	)
}

// CircuitBreakerIsOpenError is a error type for open state
type CircuitBreakerIsOpenError struct {
	Name      string
	ExpiresAt time.Time
}

func (e *CircuitBreakerIsOpenError) Error() string {
	return fmt.Sprintf(
		"circuit breaker(%s) is open, will transition to half-open at %+v",
		e.Name,
		e.ExpiresAt,
	)
}

// TimeoutError is a error type for invocation timeouts
type TimeoutError struct {
	Name     string
	Duration time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf(
		"invocation timeout for %s circuit breaker on %s",
		e.Name,
		e.Duration,
	)
}

// run executes the operator without any restriction
func (cb *CircuitBreaker) run(ctx context.Context, o Operator) (interface{}, error) {
	s := cb.State()
	if s.isClose() {
		return cb.runClose(ctx, o)
	}

	if s.isHalfOpen() {
		return cb.runHalfOpen(ctx, o)
	}

	return cb.runOpen()
}

func (cb *CircuitBreaker) runWithTimeout(ctx context.Context, o Operator) (interface{}, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, cb.invocationTimeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, &TimeoutError{Name: cb.name, Duration: cb.invocationTimeout}
	case i := <-cb.invoke(ctx, o):
		return i.res, i.err
	}
}

func (cb *CircuitBreaker) invoke(ctx context.Context, o Operator) chan invocation {
	// allow putting one invocation result into chan even if noone reads
	ch := make(chan invocation, 1)

	defer func() {
		go func() {
			// close the chan right after putting the val into it, since the
			// receive happens earlier then the put operation it won't cause any
			// problem
			defer close(ch)

			// operator can cancel execution with context timeout too
			res, err := o.Execute(ctx)

			// even if noone reads, it is non-blocking
			ch <- invocation{res: res, err: err}
		}()
	}()

	return ch
}

func (cb *CircuitBreaker) runClose(ctx context.Context, o Operator) (interface{}, error) {
	res, err := cb.runWithTimeout(ctx, o)
	if err != nil {
		failures := atomic.AddInt64(&cb.failure, 1)
		successes := atomic.LoadInt64(&cb.success)
		requests := successes + failures
		ratio := (float64(failures) / float64(requests)) * 100
		if cb.failureThreshold < ratio && cb.failureMinRequests <= requests {
			cb.open(err)
		}
		cb.runOnFailureCallbacks(StateClose, err)
		return nil, err
	}

	atomic.AddInt64(&cb.success, 1)
	cb.runOnSuccessCallbacks(res)
	return res, err
}

func (cb *CircuitBreaker) runHalfOpen(ctx context.Context, o Operator) (interface{}, error) {
	res, err := cb.runWithTimeout(ctx, o)
	if err != nil {
		cb.open(err)
		cb.runOnFailureCallbacks(StateHalfOpen, err)
		return nil, err
	}

	if cb.successThreshold <= atomic.AddInt64(&cb.success, 1) {
		cb.close()
	}
	cb.runOnSuccessCallbacks(res)
	return res, err
}

func (cb *CircuitBreaker) runOpen() (interface{}, error) {
	err := &CircuitBreakerIsOpenError{
		Name:      cb.name,
		ExpiresAt: cb.resetAt,
	}
	cb.runOnFailureCallbacks(StateOpen, err)
	return nil, err
}

func (cb *CircuitBreaker) runReject(err error) (interface{}, error) {
	cb.runOnFailureCallbacks(cb.State(), err)
	return nil, err
}

func (cb *CircuitBreaker) close() {
	if from, ok := cb.tryToClose(); ok {
		cb.runStateChangeCallbacks(from, StateClose)
	}
}

func (cb *CircuitBreaker) halfOpen() {
	if from, ok := cb.tryToHalfOpen(); ok {
		cb.runStateChangeCallbacks(from, StateHalfOpen)
	}
}

func (cb *CircuitBreaker) open(err error) {
	if from, ok := cb.tryToOpen(); ok {
		duration := cb.resetTimer.Next(err)
		cb.resetAt = time.Now().Add(duration)
		cb.runStateChangeCallbacks(from, StateOpen)
		time.AfterFunc(duration, cb.halfOpen)
	}
}

func (cb *CircuitBreaker) tryToClose() (State, bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	s := cb.state
	if s.isClose() {
		return s, false
	}

	cb.success = 0
	cb.failure = 0
	cb.state = StateClose
	cb.resetTimer.Reset()
	return s, true
}

func (cb *CircuitBreaker) tryToHalfOpen() (State, bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	s := cb.state
	if s.isHalfOpen() {
		return s, false
	}

	cb.success = 0
	cb.state = StateHalfOpen
	return s, true
}

func (cb *CircuitBreaker) tryToOpen() (State, bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	s := cb.state
	if s.isOpen() {
		return s, false
	}

	cb.state = StateOpen
	return s, true
}

func (cb *CircuitBreaker) runOnSuccessCallbacks(res interface{}) {
	for _, h := range cb.onSuccessHandlers {
		h.Handle(res)
	}
}

func (cb *CircuitBreaker) runOnFailureCallbacks(s State, err error) {
	for _, h := range cb.onFailureHandlers {
		h.Handle(s, err)
	}
}

func (cb *CircuitBreaker) runStateChangeCallbacks(from, to State) {
	for _, h := range cb.onStateChangeHandlers {
		h.Handle(from, to)
	}
}
