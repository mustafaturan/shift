// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

const (
	metricSuccess = "success"
	metricFailure = "failure"
	metricTimeout = "timeout"
	metricReject  = "reject"
)

// Stats is a structure which holds cb invocation metrics
type Stats struct {
	SuccessCount, FailureCount, TimeoutCount, RejectCount uint32
}

// newStats inits a new stats from given map
func newStats(metrics map[string]uint32) Stats {
	return Stats{
		SuccessCount: metrics[metricSuccess],
		FailureCount: metrics[metricFailure],
		TimeoutCount: metrics[metricTimeout],
		RejectCount:  metrics[metricReject],
	}
}
