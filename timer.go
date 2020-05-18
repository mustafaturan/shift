// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import "time"

// Timer is an interface to set reset time duration dynamically depending on
// the occurred error on the invocation
type Timer interface {
	// Next returns the current duration and sets the next duration according to
	// the given error
	Next(error) time.Duration

	// Reset resets the current duration to the initial duration
	Reset()
}
