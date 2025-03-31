package params

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jasonhancock/jasongen/errors"
)

type missingQueryParamErr struct {
	name string
}

func (e *missingQueryParamErr) Error() string {
	return fmt.Sprintf("query parameter %q not set", e.name)
}

func (e *missingQueryParamErr) StatusCode() int {
	return http.StatusBadRequest
}

func QueryParamBool(values url.Values, name string, dest *bool, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		boolVal, err := strconv.ParseBool(val[0])
		if err != nil {
			return errors.NewHTTP(err, http.StatusBadRequest)
		}
		*dest = boolVal
	}

	return nil
}

func QueryParamString(values url.Values, name string, dest *string, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		*dest = val[0]
	}

	return nil
}

func QueryParamInt8(values url.Values, name string, dest *int8, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 8)
		if err != nil {
			return errors.NewHTTP(err, http.StatusBadRequest)
		}
		*dest = int8(intVal)
	}

	return nil
}

func QueryParamInt16(values url.Values, name string, dest *int16, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 16)
		if err != nil {
			return errors.NewHTTP(err, http.StatusBadRequest)
		}
		*dest = int16(intVal)
	}

	return nil
}

func QueryParamInt32(values url.Values, name string, dest *int32, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 32)
		if err != nil {
			return errors.NewHTTP(err, http.StatusBadRequest)
		}
		*dest = int32(intVal)
	}

	return nil
}

func QueryParamInt64(values url.Values, name string, dest *int64, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 64)
		if err != nil {
			return errors.NewHTTP(err, http.StatusBadRequest)
		}
		*dest = intVal
	}

	return nil
}

type options struct {
	required bool
}

// Option is used to customize
type Option func(*options)

func Required(required bool) Option {
	return func(o *options) {
		o.required = required
	}
}
