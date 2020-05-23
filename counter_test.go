package shift

import (
	"testing"

	"github.com/mustafaturan/shift/counter"
)

func TestCounter(t *testing.T) {
	// Ensure TimeBucketCounter implements Counter on build
	var _ Counter = (*counter.TimeBucketCounter)(nil)
}
