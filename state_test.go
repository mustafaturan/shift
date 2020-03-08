package shift

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsClose(t *testing.T) {
	tests := []struct {
		actual   State
		expected bool
	}{
		{actual: StateClose, expected: true},
		{actual: StateHalfOpen, expected: false},
		{actual: StateOpen, expected: false},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual.isClose())
	}
}

func TestIsHalfOpen(t *testing.T) {
	tests := []struct {
		actual   State
		expected bool
	}{
		{actual: StateClose, expected: false},
		{actual: StateHalfOpen, expected: true},
		{actual: StateOpen, expected: false},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual.isHalfOpen())
	}
}

func TestIsOpen(t *testing.T) {
	tests := []struct {
		actual   State
		expected bool
	}{
		{actual: StateClose, expected: false},
		{actual: StateHalfOpen, expected: false},
		{actual: StateOpen, expected: true},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual.isOpen())
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		actual   State
		expected string
	}{
		{actual: StateClose, expected: "close"},
		{actual: StateHalfOpen, expected: "half-open"},
		{actual: StateOpen, expected: "open"},
		{actual: State(int8(-1)), expected: "unknown"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.actual.String())
	}
}
