package endpoint

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = microerror.New("invalid config")
var wrongTypeError = microerror.New("wrong type")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
