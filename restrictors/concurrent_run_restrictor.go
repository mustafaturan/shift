// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package restrictors

import (
	"fmt"
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
func (r *ConcurrentRunRestrictor) Check() (bool, error) {
	if r.threshold < atomic.AddInt64(&r.current, 1) {
		return false, &ConcurrentRunRestrictorThresholdError{
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

// ConcurrentRunRestrictorThresholdError is a error type for max concurrent runs
type ConcurrentRunRestrictorThresholdError struct {
	Name      string
	Threshold int64
}

func (e *ConcurrentRunRestrictorThresholdError) Error() string {
	return fmt.Sprintf(
		"concurrent run restriction(%s) threshold reached / runs: %d",
		e.Name,
		e.Threshold,
	)
}

// InvalidOptionError is a error tyoe for options
type InvalidOptionError struct {
	Name string
	Type string
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf(
		"invalid option provided for %s, must be %s",
		e.Name,
		e.Type,
	)
}
