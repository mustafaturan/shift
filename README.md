# 🎚Shift

[![GoDoc](https://godoc.org/github.com/mustafaturan/shift?status.svg)](https://godoc.org/github.com/mustafaturan/shift)
[![Build Status](https://travis-ci.org/mustafaturan/shift.svg?branch=master)](https://travis-ci.org/mustafaturan/shift)
[![Coverage Status](https://coveralls.io/repos/github/mustafaturan/shift/badge.svg?branch=master)](https://coveralls.io/github/mustafaturan/shift?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mustafaturan/shift)](https://goreportcard.com/report/github.com/mustafaturan/shift)
[![GitHub license](https://img.shields.io/github/license/mustafaturan/shift.svg)](https://github.com/mustafaturan/shift/blob/master/LICENSE)

Shift package is an optioned circuit breaker implementation. If you are
interested in learning what is circuit breaker or sample use cases for circuit
breakers then you can refer to [References](#References) section of this
`README.md` file.

## API

Although no significant change should be expected on design and features, the
method names and arities/args may subject to change until version `1.0.0`.
Shift package follows [semantic versioning 2.x rules](https://semver.org/) on
releases and tags. To access to the current package version, `shift.Version`
constant can be used.

## Features

* Configurable with optioned plugable components
* Comes with built-in execution timeout feature which cancels the execution by
optioned timeout duration
* Allows subscribing state change, failure and success events
* Allows overriding the current state with callbacks
* Allows adding reset timer which can be implemented using an exponential
backoff algorithm or any other algorithm when needed
* Allows adding optional restrictors by execution like max concurrent runs

## Installation

Via go packages:

```go get github.com/mustafafuran/shift```

## Usage

### Basic with defaults

#### Execute with a function implementation

```go
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
	"github.com/mustafafuran/shift/restrictors"
)

func NewCircuitBreaker() *shift.CircuitBreaker {
	restrictor := restrictors.NewConcurrentRunRestrictor("concurrent_runs", 100)
	cb, err := shift.NewCircuitBreaker(
		"twitter-cli",
		shift.WithRestrictors(restrictor),
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
	"github.com/mustafafuran/shift/timers"
)

func NewCircuitBreaker() *shift.CircuitBreaker {
	timer := timers.NewConstantTimer(5 * time.Second)
	cb, err := shift.NewCircuitBreaker(
		"twitter-cli",
		shift.WithResetTimer(timer),
		// ... other options
	)
	if err != nil {
		panic(err)
	}
	return cb
}
```

### Monitoring

Monitoring is set with options in shift package CircuitBreaker initialization.
Shift package allows adding multiple hooks on three circuit breaker events;

* **State Change Event:** Allows attaching handlers on the circuit breaker
state changes
* **Failure Event:** Allows attaching handlers on the circuit breaker
execution results with an error
* **Success Event:** Allows attaching handlers on the circuit breaker
execution results without an error

#### Configure with On State Change Handlers

```go
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
	shift.WithOnStateChangeHandlers(printer, another),
	// ... other options
)
```

#### Configure with On Failure Handlers

```go
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
	shift.WithOnFailureHandlers(printer, another, yetAnother),
	// ... other options
)
```

#### Configure with On Success Handlers

```go
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
	shift.WithOnSuccessHandlers(printer, another),
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
