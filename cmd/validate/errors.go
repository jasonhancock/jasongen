package validate

import "fmt"

var _ error = (*endpointError)(nil)

type endpointError struct {
	method string
	path   string
	err    error
}

func newEndpointError(method, path string, err error) *endpointError {
	return &endpointError{
		method: method,
		path:   path,
		err:    err,
	}
}

func (e *endpointError) Error() string {
	return fmt.Sprintf("%s %s: %s", e.method, e.path, e.err.Error())
}

func (e *endpointError) Unwrap() error {
	return e.err
}

type statusCodeMissingError struct {
	status int
}

func newStatusCodeMissingError(code int) *statusCodeMissingError {
	return &statusCodeMissingError{status: code}
}

func (e *statusCodeMissingError) Error() string {
	return fmt.Sprintf("%d response not defined", e.status)
}

type responseBodyRequiredError struct {
	status int
}

func newResponseBodyRequired(code int) *responseBodyRequiredError {
	return &responseBodyRequiredError{code}
}

func (e *responseBodyRequiredError) Error() string {
	return fmt.Sprintf("status %d requires a response body", e.status)
}
