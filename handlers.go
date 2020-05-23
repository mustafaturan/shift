// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"context"
)

// FailureHandler is an interface to handle failure events
type FailureHandler interface {
	Handle(context.Context, error)
}

// OnFailure is a function to run as a callback on any error like timeout and
// invocation errors
type OnFailure func(context.Context, error)

// Handle implements FailureHandler for OnFailure func
func (fn OnFailure) Handle(ctx context.Context, err error) {
	fn(ctx, err)
}

// SuccessHandler is an interface to handle success events
type SuccessHandler interface {
	Handle(context.Context, interface{})
}

// OnSuccess is a function to run on any successful invocation
type OnSuccess func(context.Context, interface{})

// Handle implements SuccessHandler for OnSuccess func
func (fn OnSuccess) Handle(ctx context.Context, res interface{}) {
	fn(ctx, res)
}

// StateChangeHandler is an interface to handle state change events
type StateChangeHandler interface {
	Handle(from, to State, stats Stats)
}

// OnStateChange is a function to run on any state changes
type OnStateChange func(from, to State, stats Stats)

// Handle implements StateChangeHandler for OnStateChange func
func (fn OnStateChange) Handle(from, to State, stats Stats) {
	fn(from, to, stats)
}
