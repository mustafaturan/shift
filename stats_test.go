package shift

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStats(t *testing.T) {
	metrics := map[string]uint32{
		metricSuccess: 100,
		metricFailure: 5,
		metricTimeout: 3,
		metricReject:  2,
	}

	stats := newStats(metrics)
	assert.Equal(t, metrics[metricSuccess], stats.SuccessCount)
	assert.Equal(t, metrics[metricFailure], stats.FailureCount)
	assert.Equal(t, metrics[metricTimeout], stats.TimeoutCount)
	assert.Equal(t, metrics[metricReject], stats.RejectCount)
}
