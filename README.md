# 🎚Shift

[![GoDoc](https://godoc.org/github.com/mustafaturan/shift?status.svg)](https://godoc.org/github.com/mustafaturan/shift)
[![Build Status](https://travis-ci.org/mustafaturan/shift.svg?branch=master)](https://travis-ci.org/mustafaturan/shift)
[![Coverage Status](https://coveralls.io/repos/github/mustafaturan/shift/badge.svg?branch=master)](https://coveralls.io/github/mustafaturan/shift?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mustafaturan/shift)](https://goreportcard.com/report/github.com/mustafaturan/shift)
[![GitHub license](https://img.shields.io/github/license/mustafaturan/shift.svg)](https://github.com/mustafaturan/shift/blob/master/LICENSE)

Shift package is an optioned circuit breaker implementation.

**For those who are new to the concept, a brief summary:**
* circuit breaker has 3 states: `close`, `half-open` and `open`
* when it is in `open` state, *something bad is going on the executions* and to
prevent bad invokations, the circuit breaker gave a break to new invokations,
and returns error
* when it is in `half-open` state, then there *could be a chance for recovery*
from bad state and circuit breaker evaluates criterias to trip to next states
* when it is in `close` state, then *everything is working as expected*

**State changes in circuit breaker:**
* `close -> open`: `close` can trip to `open` state
* `close <- half-open -> open`: `half-open` can trip to both `close` and `open`
states
* `open -> half-open`: `open` can trip to `half-open`

If you are interested in learning what is circuit breaker or sample use cases
for circuit breakers then you can refer to [References](#References) section of
this `README.md` file.

## API

Shift package follows [semantic versioning 2.x rules](https://semver.org/) on
releases and tags. To access to the current package version, `shift.Version`
constant can be used.

## Features

* Configurable with optioned plugable components
* Comes with built-in execution timeout feature which cancels the execution by
optioned timeout duration
* Comes with built-in bucketted counter feature which counts the stats by given
durationed buckets
* Allows subscribing state change, failure and success events
* Allows overriding the current state with callbacks
* Allows overriding reset timer which can be implemented using an exponential
backoff algorithm or any other algorithm when needed
* Allows overriding counter which can allow using an external counter for
managing the stats
* Allows adding optional restrictors by execution like max concurrent runs

## Installation

Via go packages:

```go get github.com/mustafafuran/shift```

## Usage

### Basic with defaults

On configurations 3 options are critical to have a healthy circuit breaker,
so on any configuration it is highly recommended to specify at least the
following 3 options with desired numbers.

```go
// Trip from Close to Open state under 95.0% success ratio at minimum of
// 20 invokations on configured duration(see WithCounter for durationed stats)
shift.WithOpener(StateClose, 95.0, 20),

// Trip from Half-Open to Open state under 75.0% success ratio at minimum of
// 10 invokations on configured duration(see WithCounter for durationed stats)
shift.WithOpener(StateHalfOpen, 75.0, 10),

// Trip from Half-Open to Close state on 90.0% of success ratio at minumum of
// 10 invokations on configured duration(see WithCounter for durationed stats)
shift.WithCloser(90.0, 10),
```

It is also recommended to have multiple circuit breakers for each invoker with
their own specific configurations depending on the SLA requirements.
For example: A circuit breaker for Github, another for Twitter, and yet
another for Facebook API client.

#### Execute with a function implementation

```go
import (
	"context"
	"fmt"

	"github.com/mustafafuran/shift"
)

func NewCircuitBreaker() *shift.Shift {
	cb, err := shift.New(
		"a-name-for-the-breaker",

		// Trip from Close to Open state under 95.0% success ratio at minimum of
		// 20 invokations on configured duration(see WithCounter for durationed
		// stats)
		shift.WithOpener(StateClose, 95.0, 20),

		// Trip from Half-Open to Open state under 75.0% success ratio at
		// minimum of 10 invokations on configured duration(see WithCounter for
		// durationed stats)
		shift.WithOpener(StateHalfOpen, 75.0, 10),

		// Trip from Half-Open to Close state on 90.0% of success ratio at
		// minumum of 10 invokations on configured duration(see WithCounter for
		// durationed stats)
		shift.WithCloser(90.0, 10),
	)
	if err != nil {
		panic(err)
	}
	return cb
}

func DoSomethingWithFn(ctx context.Context, cb *shift.CircuitBreaker) string {
	var fn shift.Operate = func(ctx context.Context) (interface{}, error) {
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
```

### Configure for max concurrent runnables

Shift allows adding restrictors like max concurrent runnables to prevent
execution of the `invokes` on developer defined conditions. Restrictors do not
effect the current state, but they can block the execution depending on their
own internal state values. If a restrictor blocks an execution then it returns
an error and `On Failure Handlers` get executed in order.

```go
import (
	"github.com/mustafafuran/shift"
	"github.com/mustafafuran/shift/restrictor"
)

func NewCircuitBreaker() *shift.CircuitBreaker {
	restrictor, err := restrictor.NewConcurrentRunRestrictor("concurrent_runs", 100)
	if err != nil {
		return err
	}

	cb, err := shift.New(
		"twitter-cli",

		// Restrictors
		shift.WithRestrictors(restrictor),

		// Trippers
		shift.WithOpener(StateClose, 95.0, 20),
		shift.WithOpener(StateHalfOpen, 75.0, 10),
		shift.WithCloser(90.0, 10),

		// ... other options
	)
	if err != nil {
		panic(err)
	}
	return cb
}
```

### Creating a reset timer based on errors

Any reset timer strategy can be implemented on top of `shift.Timer` interface.
The default timer strategy does not intentionally implement any use case
specific strategy like exponential backoff. Since the decision of reset time
incrementation should be taken depending on error reasons, the best decider for
each instance of CircuitBreaker would be the developers. In case, if it is good
to just have a constant timeout duration, the `shift/timer.ConstantTimer`
implementation should simply help to configure your reset timeout duration.

```go

import (
	"github.com/mustafafuran/shift"
	"github.com/mustafafuran/shift/timer"
)

func NewCircuitBreaker() *shift.CircuitBreaker {
	timer, err := timer.NewConstantTimer(5 * time.Second)
	if err != nil {
		panic(err)
	}

	cb, err := shift.New(
		"twitter-cli",
		// Reset Timer
		shift.WithResetTimer(timer),

		// Trippers
		shift.WithOpener(StateClose, 95.0, 20),
		shift.WithOpener(StateHalfOpen, 75.0, 10),
		shift.WithCloser(90.0, 10),

		// ... other options
	)
	if err != nil {
		panic(err)
	}
	return cb
}
```

### Creating a counter based on your bucketing needs

Any counter strategy can be implemented on top of `shift.Counter` interface.
The default counter strategy is using a bucketing mechanism to bucket time and
add/drop metrics into the stats. The default Counter uses 1 second durationed
10 buckets. There are two possible options to modify the Counter based on your
needs:

1) Create a new counter instance and pass as counter option

2) Create your own counter implementations and pass the instance as counter
option

```go

import (
	"github.com/mustafafuran/shift"
	"github.com/mustafafuran/shift/counter"
)

func NewCircuitBreaker() *shift.CircuitBreaker {
	// The TimeBucketCounter automatically drops a the oldest bucket after
	// filling the available last bucket and then shift left the buckets, so a
	// new space is freeing up for a new bucket

	// 60 buckets each holds the stats for 2 seconds
	capacity, duration := 60, 2000 * time.Millisecond
	counter, err := counter.TimeBucketCounter(capacity, duration)
	if err != nil {
		panic(err)
	}

	cb, err := shift.New(
		"twitter-cli",
		// Counter
		shift.WithCounter(counter),

		// Trippers
		shift.WithOpener(StateClose, 95.0, 20),
		shift.WithOpener(StateHalfOpen, 75.0, 10),
		shift.WithCloser(90.0, 10),

		// ... other options
	)
	if err != nil {
		panic(err)
	}
	return cb
}
```

### Events

Shift package allows adding multiple hooks on failure, success and state change
circuit breaker events. Both success and failure events come with a context
which holds state and stats;

* **State Change Event:** Allows attaching handlers on the circuit breaker
state changes
* **Failure Event:** Allows attaching handlers on the circuit breaker
execution results with an error
* **Success Event:** Allows attaching handlers on the circuit breaker
execution results without an error

#### Configure with On State Change Handlers

```go
// a printer handler
var printer shift.OnStateChange = func(from, to shift.State, stats shift.Stats) {
	fmt.Printf("State changed from %s, to %s, %+v", from, to, stats)
}

// another handler
var another shift.OnStateChange = func(from, to shift.State, stats shift.Stats) {
	// do sth
}

cb, err := shift.New(
	"a-name",
	shift.WithStateChangeHandlers(printer, another),
	// ... other options
)
```

#### Configure with On Failure Handlers

```go
// a printer handler
var printer shift.OnFailure = func(ctx context.Context, err error) {
	state := ctx.Value(CtxState).(State)
	stats := ctx.Value(CtxStats).(Stats)

	fmt.Printf("execution erred on state(%s) with %s and stats are %+v", state, err, stats)
}

// another handler
var another shift.OnFailure = func(ctx context.Context, err error) {
	// do sth: maybe increment an external metric when the execution err
}

// yetAnother handler
var yetAnother shift.OnFailure = func(ctx context.Context, err error) {
	// do sth
}

cb, err := shift.New(
	"a-name",
	// appends the failure handlers provided
	shift.WithFailureHandlers(StateClose, printer, another, yetAnother),
	shift.WithFailureHandlers(StateHalfOpen, printer, another),

	// Trippers
	shift.WithOpener(StateClose, 95.0, 20),
	shift.WithOpener(StateHalfOpen, 75.0, 10),
	shift.WithCloser(90.0, 10),

	// ... other options
)
```

#### Configure with On Success Handlers

```go
// a printer handler
var printer shift.OnSuccess = func(ctx context.Context, data interface{}) {
	state := ctx.Value(CtxState).(State)
	stats := ctx.Value(CtxStats).(Stats)

	fmt.Printf("execution succeeded on %s and resulted with %+v and stats are %+v", state, data, stats)
}

// another handler
var another shift.OnSuccess = func(ctx context.Context, data interface{}) {
	// do sth: maybe increment an external metric when the execution succeeds
}

cb, err := shift.New(
	"a-name",

	// Appends the success handlers for a given state
	shift.WithSuccessHandlers(StateClose, printer, another),
	shift.WithSuccessHandlers(StateHalfOpen, printer),

	// Trippers
	shift.WithOpener(StateClose, 95.0, 20),
	shift.WithOpener(StateHalfOpen, 75.0, 10),
	shift.WithCloser(90.0, 10),

	// ... other options
)
```

#### Advanced configuration options

Please refer to [GoDoc](https://godoc.org/github.com/mustafaturan/shift) for
more options and configurations.

#### Examples

[Shift Examples Repository](https://github.com/mustafaturan/shift-examples)

## Contributing

All contributors should follow [Contributing Guidelines](CONTRIBUTING.md) before
creating pull requests.

### New features

All features SHOULD be optional and SHOULD NOT change the API contracts. Please
refer to [API](#API) section of the [README.md](README.md) for more information.

### Unit tests

Test coverage is very important for this kind of important piece of
infrastructure softwares. Any change MUST cover all use cases with race
condition checks.

To run unit tests locally, you can use `Makefile` short cuts:

```bash
make test_race # test against race conditions
make test # test and write the coverage results to `./coverage.out` file
make coverage # display the coverage in format
make all # run all three above in order
```

## References

* [Microsoft Docs - Circuit Breaker Pattern](https://docs.microsoft.com/en-us/previous-versions/msp-n-p/dn589784(v=pandp.10))
* [Martin Fowler - Circuit Breaker](https://martinfowler.com/bliki/CircuitBreaker.html)
* [Netflix - Hystrix(Java)](https://github.com/Netflix/Hystrix/)
* [Hystrix-Go](https://github.com/afex/hystrix-go)
* [Sony - GoBreaker](https://github.com/sony/gobreaker)

## Credits

[Mustafa Turan](https://github.com/mustafaturan)

## License

Apache License 2.0

Copyright (c) 2020 Mustafa Turan

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
