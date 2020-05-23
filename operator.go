// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import "context"

// Operator is an interface for circuit breaker operations
type Operator interface {
	Execute(context.Context) (interface{}, error)
}

// Operate is a function that runs the operation
type Operate func(context.Context) (interface{}, error)

// Execute implements Operator interface for any Operate function for free
func (o Operate) Execute(ctx context.Context) (interface{}, error) {
	return o(ctx)
}
