package shift

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperate(t *testing.T) {
	// Ensure Operate implements Operator on build
	var _ Operator = (Operate)(nil)

	var fn Operate = func(context.Context) (interface{}, error) {
		return true, nil
	}

	res, err := fn.Execute(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, true, res.(bool))
}
