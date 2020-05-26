package shift

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeadlineInvoker_Invoke(t *testing.T) {
	t.Run("with timeout", func(t *testing.T) {
		var called bool
		invoker := &deadlineInvoker{
			timeout:         time.Millisecond,
			timeoutCallback: func() { called = true },
		}

		var fn Operate = func(context.Context) (interface{}, error) {
			time.Sleep(2 * time.Millisecond)
			return nil, nil
		}
		res, err := invoker.invoke(context.Background(), fn)

		assert.Error(t, err)
		assert.IsType(t, &InvocationTimeoutError{}, err)
		assert.Nil(t, res)
		assert.Equal(t, true, called)
	})

	t.Run("without timeout", func(t *testing.T) {
		var called bool
		invoker := &deadlineInvoker{
			timeout:         time.Second,
			timeoutCallback: func() { called = true },
		}

		t.Run("on failure", func(t *testing.T) {
			var fn Operate = func(context.Context) (interface{}, error) {
				return nil, errors.New("operation error")
			}
			res, err := invoker.invoke(context.Background(), fn)

			assert.Error(t, err)
			assert.EqualError(t, err, "operation error")
			assert.Nil(t, res)
			assert.Equal(t, false, called)
		})

		t.Run("on success", func(t *testing.T) {
			const val = "test"
			var fn Operate = func(context.Context) (interface{}, error) {
				return val, nil
			}
			res, err := invoker.invoke(context.Background(), fn)

			assert.NoError(t, err)
			assert.Equal(t, val, res.(string))
			assert.Equal(t, false, called)
		})
	})
}

func TestOnOpenInvoker_Invoke(t *testing.T) {
	var called bool
	invoker := &onOpenInvoker{rejectCallback: func() {
		called = true
	}}

	var fn Operate = func(context.Context) (interface{}, error) { return nil, nil }
	res, err := invoker.invoke(context.Background(), fn)

	assert.Error(t, err)
	assert.IsType(t, &IsOnOpenStateError{}, err)
	assert.Nil(t, res)
	assert.Equal(t, true, called)
}
