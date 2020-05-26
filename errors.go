// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"fmt"
	"time"
)

// InvalidOptionError is a error tyoe for options
type InvalidOptionError struct {
	Name    string
	Message string
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf(
		"invalid option provided for %s: %s",
		e.Name,
		e.Message,
	)
}

// UnknownStateError is a error tyoe for states
type UnknownStateError struct {
	State State
}

func (e *UnknownStateError) Error() string {
	return fmt.Sprintf(
		"unknown state(%d) provided, the allowed states are 'close', 'half-open' and 'open'",
		e.State,
	)
}

// IsAlreadyInDesiredStateError is an error type for stating the
// current state is already in desired state
type IsAlreadyInDesiredStateError struct {
	Name  string
	State State
}

func (e *IsAlreadyInDesiredStateError) Error() string {
	return fmt.Sprintf(
		"circuit breaker(%s) is already in the desired state(%s)",
		e.Name,
		e.State,
	)
}

// IsOnOpenStateError is a error type for open state
type IsOnOpenStateError struct{}

func (e *IsOnOpenStateError) Error() string {
	return "is on open state"
}

// InvocationError is an error type to wrap invocation errors
type InvocationError struct {
	Name string
	Err  error
}

func (e *InvocationError) Error() string {
	return fmt.Sprintf("circuit breaker(%s) invocation failed with %s", e.Name, e.Err)
}

func (e *InvocationError) Unwrap() error {
	return e.Err
}

// InvocationTimeoutError is a error type for invocation timeouts
type InvocationTimeoutError struct {
	Duration time.Duration
}

func (e *InvocationTimeoutError) Error() string {
	return fmt.Sprintf(
		"invocation timeout on %s",
		e.Duration,
	)
}

// FailureThresholdReachedError is a error type for failure threshold
type FailureThresholdReachedError struct{}

func (e *FailureThresholdReachedError) Error() string {
	return "failure threshold reached"
}
