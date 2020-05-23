package restrictor

import "fmt"

// InvalidOptionError is a error tyoe for options
type InvalidOptionError struct {
	Name string
	Type string
}

func (e *InvalidOptionError) Error() string {
	return fmt.Sprintf(
		"invalid option provided for %s, must be %s",
		e.Name,
		e.Type,
	)
}

// ThresholdError is a error type for max concurrent runs
type ThresholdError struct {
	Name      string
	Threshold int64
}

func (e *ThresholdError) Error() string {
	return fmt.Sprintf(
		"concurrent run restriction(%s) threshold reached / runs: %d",
		e.Name,
		e.Threshold,
	)
}
