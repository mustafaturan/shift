// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package timers

import "time"

// ConstantTimer holds contant duration and regardless of how many times
// called it always return the initiated constant duration
type ConstantTimer struct {
	duration time.Duration
}

// NewConstantTimer inits ConstantTimer with the given duration
func NewConstantTimer(d time.Duration) *ConstantTimer {
	return &ConstantTimer{duration: d}
}

// Next returns always the same duration regardless of the error type
func (c *ConstantTimer) Next(_ error) time.Duration {
	return c.duration
}

// Reset sets the current duration to the initial duration
func (c *ConstantTimer) Reset() {}
