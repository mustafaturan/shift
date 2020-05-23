package timer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidOptionError(t *testing.T) {
	err := &InvalidOptionError{
		Name: "test",
		Type: "positive duration in seconds",
	}
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid option provided for test, must be positive duration in seconds")
}
