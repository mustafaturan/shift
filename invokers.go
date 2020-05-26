// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package shift

import (
	"context"
	"time"
)

type invoker interface {
	invoke(context.Context, Operator) (interface{}, error)
}

type deadlineInvoker struct {
	timeout         time.Duration
	timeoutCallback func()
}

type onCloseInvoker = deadlineInvoker

type onHalfOpenInvoker = deadlineInvoker

type onOpenInvoker struct {
	rejectCallback func()
}

// invocation is a type for holding invocation result
type invocation struct {
	res interface{}
	err error
}

/* on open state */

func (i *onOpenInvoker) invoke(ctx context.Context, o Operator) (interface{}, error) {
	i.rejectCallback()
	return nil, &IsOnOpenStateError{}
}

/* on half-open & close states */

func (i *deadlineInvoker) invoke(ctx context.Context, o Operator) (interface{}, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, i.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		i.timeoutCallback()
		return nil, &InvocationTimeoutError{Duration: i.timeout}
	case i := <-i.async(ctx, o):
		return i.res, i.err
	}
}

func (i *deadlineInvoker) async(ctx context.Context, o Operator) chan invocation {
	// allow putting one invocation result into chan even if noone reads
	ch := make(chan invocation, 1)

	defer func() {
		go func() {
			// close the chan right after putting the val into it, since the
			// receive happens earlier then the put operation it won't cause any
			// problem
			defer close(ch)

			// operator can cancel execution with context timeout too
			res, err := o.Execute(ctx)

			// even if noone reads, it is non-blocking with the buffered channel
			ch <- invocation{res: res, err: err}
		}()
	}()

	return ch
}
