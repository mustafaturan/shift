package shift

import (
	"testing"

	"github.com/mustafaturan/shift/restrictor"
)

func TestRestrictor(t *testing.T) {
	var _ Restrictor = (*restrictor.ConcurrentRunRestrictor)(nil)
}
