package params

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jasonhancock/go-helpers"
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

func QueryParamBool(values url.Values, name string, opts ...Option) (*bool, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		boolVal, err := strconv.ParseBool(val[0])
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		return &boolVal, nil
	}

	return nil, nil
}

func QueryParamString(values url.Values, name string, opts ...Option) (*string, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		return &val[0], nil
	}

	return nil, nil
}

func QueryParamInt8(values url.Values, name string, opts ...Option) (*int8, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 8)
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		return helpers.Ptr(int8(intVal)), nil
	}

	return nil, nil
}

func QueryParamInt16(values url.Values, name string, opts ...Option) (*int16, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 16)
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		return helpers.Ptr(int16(intVal)), nil
	}

	return nil, nil
}

func QueryParamInt32(values url.Values, name string, opts ...Option) (*int32, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 32)
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		return helpers.Ptr(int32(intVal)), nil
	}

	return nil, nil
}

func QueryParamInt64(values url.Values, name string, opts ...Option) (*int64, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingQueryParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 64)
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		return &intVal, nil
	}

	return nil, nil
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
