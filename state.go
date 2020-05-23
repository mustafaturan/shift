// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

// State circuit breaker state holder
type State int8

const (
	// StateUnknown is an unknown state for circuit breaker
	StateUnknown State = iota
	// StateClose close state for circuit breaker
	StateClose
	// StateHalfOpen half-open state for circuit breaker
	StateHalfOpen
	// StateOpen open state for circuit breaker
	StateOpen
)

func (s State) isClose() bool {
	return s == StateClose
}

func (s State) isHalfOpen() bool {
	return s == StateHalfOpen
}

func (s State) isOpen() bool {
	return s == StateOpen
}

func (s State) String() string {
	switch s {
	case StateClose:
		return "close"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}
