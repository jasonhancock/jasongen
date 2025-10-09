package params

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jasonhancock/go-helpers"
	"github.com/jasonhancock/jasongen/errors"
)

type missingParamErr struct {
	name string
}

func (e *missingParamErr) Error() string {
	return fmt.Sprintf("query parameter %q not set", e.name)
}

func (e *missingParamErr) StatusCode() int {
	return http.StatusBadRequest
}

func paramBool(values map[string][]string, name string, opts ...Option) (*bool, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
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

func paramString(values map[string][]string, name string, opts ...Option) (*string, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		return &val[0], nil
	}

	return nil, nil
}

func paramInt8(values url.Values, name string, opts ...Option) (*int8, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
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

func paramInt16(values url.Values, name string, opts ...Option) (*int16, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
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

func paramInt32(values url.Values, name string, opts ...Option) (*int32, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
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

func paramInt64(values url.Values, name string, opts ...Option) (*int64, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
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

func paramFloat32(values url.Values, name string, opts ...Option) (*float32, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		fltVal, err := strconv.ParseFloat(val[0], 32)
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		flt32 := float32(fltVal)
		return &flt32, nil
	}

	return nil, nil
}

func paramFloat64(values url.Values, name string, opts ...Option) (*float64, error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		return nil, &missingParamErr{name}
	}

	if o.required || (ok && val[0] != "") {
		fltVal, err := strconv.ParseFloat(val[0], 64)
		if err != nil {
			return nil, errors.NewHTTP(err, http.StatusBadRequest)
		}
		return &fltVal, nil
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
