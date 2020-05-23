// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"context"
	"time"
)

type ctxKey string

const (
	// CtxState holds state context key
	CtxState = ctxKey("state")

	// CtxStats holds stats context key
	CtxStats = ctxKey("stats")
)

// Run executes the given func with circuit breaker
func (s *Shift) Run(ctx context.Context, o Operator) (interface{}, error) {
	ctx = context.WithValue(ctx, CtxState, s.currentState())
	return s.runWithCallbacks(ctx, o)
}

// Trip to desired state
func (s *Shift) Trip(to State, reasons ...error) error {
	stats := s.stats()

	from, err := s.trip(to, reasons...)
	if err != nil {
		return err
	}

	s.runStateChangeCallbacks(from, to, stats)
	return nil
}

func (s *Shift) trip(to State, reasons ...error) (State, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state := s.state
	if state == to {
		return state, &IsAlreadyInDesiredStateError{
			Name:  s.name,
			State: state,
		}
	}

	switch to {
	case StateClose:
		s.close()
	case StateHalfOpen:
		s.halfOpen()
	case StateOpen:
		var reason error
		if len(reasons) > 0 {
			reason = reasons[0]
		}
		s.open(reason)
	default:
		return state, &UnknownStateError{State: to}
	}

	return state, nil
}

// Close the circuit breaker
func (s *Shift) close() {
	// Set state
	s.state = StateClose

	// Reset timer
	s.resetTimer.Reset()

	// Reset counter
	s.counter.Reset()
}

// HalfOpen the circuit breaker
func (s *Shift) halfOpen() {
	// Set state
	s.state = StateHalfOpen

	// Reset counter
	s.counter.Reset()
}

// Open the circuit breaker
func (s *Shift) open(reason error) {
	// Fetch next reset duration
	duration := s.resetTimer.Next(reason)

	// Prevent a possible early switch to half-open state by stopping a possible
	// active timer
	s.resetter.Stop()

	// Reset the resetter
	s.resetter = time.AfterFunc(duration, func() {
		_ = s.Trip(StateHalfOpen)
	})

	// Set state
	s.state = StateOpen

	// Reset counter
	s.counter.Reset()
}

/* stats */

// stats returns the stats for invocations
func (s *Shift) stats() Stats {
	stats := s.counter.Stats(
		metricSuccess,
		metricFailure,
		metricTimeout,
		metricReject,
	)
	return newStats(stats)
}

/* instance accessors */

// currentState returns current state of the circuit breaker
func (s *Shift) currentState() State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state
}

/* runners */

func (s *Shift) runWithCallbacks(ctx context.Context, o Operator) (interface{}, error) {
	res, err := s.run(ctx, o)

	// Wrap the error with additional circuit breaker name information
	if err != nil {
		err = &InvokationError{Name: s.name, Err: err}
		s.runFailureCallbacks(ctx, err)
	} else {
		s.runSuccessCallbacks(ctx, res)
	}

	return res, err
}

func (s *Shift) run(ctx context.Context, o Operator) (interface{}, error) {
	for _, r := range s.restrictors {
		defer r.Defer()
		if ok, err := r.Check(ctx); !ok {
			s.counter.Increment(metricReject)
			return nil, err
		}
	}

	state := ctx.Value(CtxState).(State)
	return s.invokers[state].invoke(ctx, o)
}

/* callbacks */

func (s *Shift) runSuccessCallbacks(ctx context.Context, res interface{}) {
	s.counter.Increment(metricSuccess)

	state := ctx.Value(CtxState).(State)
	handlers := s.successHandlers[state]
	if len(handlers) == 0 {
		return
	}

	ctx = context.WithValue(ctx, CtxStats, s.stats())
	for _, h := range handlers {
		h.Handle(ctx, res)
	}
}

func (s *Shift) runFailureCallbacks(ctx context.Context, err error) {
	s.counter.Increment(metricFailure)

	state := ctx.Value(CtxState).(State)
	handlers := s.failureHandlers[state]
	if len(handlers) == 0 {
		return
	}

	ctx = context.WithValue(ctx, CtxStats, s.stats())
	for _, h := range handlers {
		h.Handle(ctx, err)
	}
}

func (s *Shift) runStateChangeCallbacks(from, to State, stats Stats) {
	handlers := s.stateChangeHandlers
	for _, h := range handlers {
		h.Handle(from, to, stats)
	}
}
