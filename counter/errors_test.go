package counter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidOptionError(t *testing.T) {
	err := &InvalidOptionError{
		Name: "test",
		Type: "uint32",
	}
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid option provided for test, must be uint32")
}
