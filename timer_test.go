package shift

import (
	"testing"

	"github.com/mustafaturan/shift/timer"
)

func TestTimer(t *testing.T) {
	// Ensure ConstantTimer implements Timer on build
	var _ Timer = (*timer.ConstantTimer)(nil)
}
