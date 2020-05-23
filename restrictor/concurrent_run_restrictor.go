// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package restrictor

import (
	"context"
	"sync/atomic"
)

// ConcurrentRunRestrictor is a restrictor for concurrent runs
type ConcurrentRunRestrictor struct {
	name               string
	current, threshold int64
}

// NewConcurrentRunRestrictor inits a new concurrent run restrictor
func NewConcurrentRunRestrictor(name string, threshold int64) (*ConcurrentRunRestrictor, error) {
	if threshold < 1 {
		return nil, &InvalidOptionError{
			Name: "concurrent run threshold",
			Type: "positive integer",
		}
	}
	return &ConcurrentRunRestrictor{
		name:      name,
		threshold: threshold,
	}, nil
}

// Check checks if possible to add new runs
func (r *ConcurrentRunRestrictor) Check(_ context.Context) (bool, error) {
	if r.threshold < atomic.AddInt64(&r.current, 1) {
		return false, &ThresholdError{
			Name:      r.name,
			Threshold: r.threshold,
		}
	}
	return true, nil
}

// Defer removes 1 from current runs
func (r *ConcurrentRunRestrictor) Defer() {
	atomic.AddInt64(&r.current, -1)
}
