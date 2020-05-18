// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

// Counter is an interface to increment, reset and fetch invocation stats
type Counter interface {
	Increment(metric string)
	Stats(metrics ...string) map[string]uint32
	Reset()
}
