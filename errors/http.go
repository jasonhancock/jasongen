package errors

// HTTPErr is a error that provides a status code.
type HTTP struct {
	error
	Code int
}

// NewHTTP generates a new HTTP.
func NewHTTP(err error, code int) *HTTP {
	return &HTTP{
		error: err,
		Code:  code,
	}
}

// Unwrap supports the Go 1.13 error semantics.
func (h *HTTP) Unwrap() error {
	return h.error
}

// StatusCode provides the status code associated with the error message.
func (h *HTTP) StatusCode() int {
	return h.Code
}
