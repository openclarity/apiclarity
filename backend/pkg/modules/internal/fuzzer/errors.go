package fuzzer

import (
	"fmt"
)

/*
*  Error used when an object does not exists on the system.
 */
type NotFoundError struct {
	msg string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Not found error: %v", e.msg)
}

/*
*  Error used when action is impossible because nor enough parameters of invalid parameter.
 */
type InvalidParameterError struct {
	msg string
}

func (e *InvalidParameterError) Error() string {
	return fmt.Sprintf("Invalid parameter error: %v", e.msg)
}

/*
*  General error.
 */
type PluginError struct {
	msg string
}

func (e *PluginError) Error() string {
	return fmt.Sprintf("General Fuzzer error: %v", e.msg)
}

/*
*  Error for an usupported action/parameter.
 */
type NotSupportedError struct {
	msg string
}

func (e *NotSupportedError) Error() string {
	return fmt.Sprintf("Not supported action or parameter: %v", e.msg)
}
