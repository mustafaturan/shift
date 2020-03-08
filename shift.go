// Copyright 2020 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

/*
Package shift is an optioned circuit breaker implementation.

Features

* Every component is optional with defaults
* Comes with built-in execution timeout feature which cancels the execution by
optioned timeout duration
* Allows subscribing state change, failure and success events
* Allows overriding the current state with callbacks
* Allows adding reset timer which can be implemented using an exponential
backoff algorithm or any other algorithm when needed
* Allows adding restrictors like max concurrent runs, and any other restrictor
can be implemented

Execute with a function implementation:

	import (
		"context"
		"fmt"

		"github.com/mustafafuran/shift"
	)

	func NewCircuitBreaker() *shift.CircuitBreaker {
		cb, err := shift.NewCircuitBreaker("a-name-for-the-breaker")
		if err != nil {
			panic(err)
		}
		return cb
	}

	func DoSomethingWithFn(ctx context.Context, cb *shift.CircuitBreaker) string {
		var fn shift.Operate = func(ctx context.Context, ...interface{}) (interface{}, error) {
			// do something in here
			return "foo", nil
		}
		res, err := cb.Run(ctx, fn)
		if err != nil {
			// maybe read from cache to set the res again?
		}

		// convert your res into your actual data
		data := res.(string)
		return data
	}

	func main() {
		cb := NewCircuitBreaker()
		data := DoSomethingWithFn(ctx, cb)
		fmt.Printf("data: %s\n", data)
	}

Execute with an interface implementation:

	import (
		"context"
		"fmt"

		"github.com/mustafafuran/shift"
	)

	func NewCircuitBreaker() *shift.CircuitBreaker {
		cb, err := shift.NewCircuitBreaker("twitter-cli")
		if err != nil {
			panic(err)
		}
		return cb
	}

	type TweetFetcher struct {
		cb *shift.CircuitBreaker
	}

	type Tweet struct {
		ID 	  int
		Title string
	}

	func (tf *TweetFetcher) Execute(ctx context.Context, opts ...interface{}) (interface{}, error) {
		var tweet *Tweet
		var err error

		id := opts[0].(int)

		// do something to fetch a tweet information
		tweet = &Tweet{
			ID:    id,
			Title: "Fake tweet",
		}
		return tweet, err
	}

	func (tf *TweetFetcher) FetchWithCircuitBreaker(ctx context.Context, id int) *Tweet {
		res, err := tf.cb.Run(ctx, tf, id)
		if err != nil {
			// maybe read from cache to set the res again?
		}

		// convert your res into your actual data
		tweet := res.(*Tweet)
		return tweet
	}

	func main() {
		cb, err := NewCircuitBreaker()
		tf := &TweetFetcher{cb}
		tweet := tf.FetchWithCircuitBreaker(ctx, 1)
		fmt.Printf("tweet: %+v\n", *tweet)
	}

Configure for max concurrent runnables

Shift allows adding restrictors like max concurrent runnables to prevent
execution of the `run` on developer defined conditions. Restrictors does not
effect the current state, but they can block the execution depending on its own
internal state values. If a restrictor blocks an execution then it returns an
error and `On Failure Handlers` get executed in order.

	import (
		"github.com/mustafafuran/shift"
		"github.com/mustafafuran/shift/restrictors"
	)

	func NewCircuitBreaker() *shift.CircuitBreaker {
		restrictor := restrictors.NewConcurrentRunRestrictor("concurrent_runs", 100)
		cb, err := shift.NewCircuitBreaker(
			"twitter-cli",
			WithRestrictors(restrictor),
			// ... other options
		)
		if err != nil {
			panic(err)
		}
		return cb
	}

Creating a reset timer based on errors

Any reset timer strategy can be implemented on top of `shift.Timer` interface.
The default timer strategy does not intentionally implement any use case
specific strategy like exponential backoff. Since the decision of reset time
incrementation should be taken depending on error reasons, the best decider for
each instance of CircuitBreaker would be the developers. In case, if it is good
to just have a constant timeout duration, the `shift/timer.ConstantTimer`
implementation should help to configure your reset timeout duration simply.

	import (
		"github.com/mustafafuran/shift"
		"github.com/mustafafuran/shift/timers"
	)

	func NewCircuitBreaker() *shift.CircuitBreaker {
		timer := timers.NewConstantTimer(5 * time.Second)
		cb, err := shift.NewCircuitBreaker(
			"twitter-cli",
			WithResetTimer(timer),
			// ... other options
		)
		if err != nil {
			panic(err)
		}
		return cb
	}

Monitoring

Monitoring is set with options in shift package CircuitBreaker initializations.
Shift package allows adding multiple hooks on three circuit breaker events;

* **State Change Event:** Allows attaching handlers when the circuit breaker
state changes
* **Failure Event:** Allows attaching handlers when the circuit breaker
execution results with an error
* **Success Event:** Allows attaccing handlers when the circuit breaker
execution results without an error

Configure with On State Change Handlers:

	// a printer handler
	var printer shift.OnStateChange = func(from, to shift.State) {
		fmt.Printf("State changed from %s, to %s", from, to)
	}

	// another handler
	var another shift.OnStateChange = func(from, to shift.State) {
		// do sth
	}

	cb, err := shift.NewCircuitBreaker(
		"a-name",
		WithOnStateChangeHandlers(printer, another),
		// ... other options
	)

Configure with On Failure Handlers:

	// a printer handler
	var printer shift.OnFailure = func(state shift.State, err error) {
		fmt.Printf("execution erred on state(%s) with %s", state, err)
	}

	// another handler
	var another shift.OnFailure = func(state shift.State, err error) {
		// do sth: maybe increment metrics when the execution err
	}

	// yetAnother handler
	var yetAnother shift.OnFailure = func(state shift.State, err error) {
		// do sth
	}

	cb, err := shift.NewCircuitBreaker(
		"a-name",
		WithOnFailureHandlers(printer, another, yetAnother),
		// ... other options
	)

Configure with On Success Handlers:

	// a printer handler
	var printer shift.OnSuccess = func(data interface{}) {
		fmt.Printf("execution succeeded and resulted with %+v", data)
	}

	// another handler
	var another shift.OnSuccess = func(data interface{}) {
		// do sth: maybe increment metrics when the execution succeeds
	}

	cb, err := shift.NewCircuitBreaker(
		"a-name",
		WithOnSuccessHandlers(printer, another),
		// ... other options
	)

*/
package shift

const (
	// Version is the current shift package version
	Version = "0.1.0"
)
