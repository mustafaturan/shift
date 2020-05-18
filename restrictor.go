// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import "context"

// Restrictor allows adding restriction to circuit breaker
type Restrictor interface {
	// Check checks if restriction allows to run current invocation and errors
	// if not allowed the invocation
	Check(context.Context) (bool, error)

	// Defer executes exit rules of the restrictor right after the run process
	Defer()
}
