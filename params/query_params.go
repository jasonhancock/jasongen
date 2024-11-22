package params

import (
	"fmt"
	"net/url"
	"strconv"
)

func QueryParamString(values url.Values, name string, dest *string, opts ...Option) error {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	val, ok := values[name]
	if o.required && !ok {
		// TODO: this should be a 400 bad request
		return fmt.Errorf("query parameter %q not set", name)
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
		// TODO: this should be a 400 bad request
		return fmt.Errorf("query parameter %q not set", name)
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 8)
		if err != nil {
			// TODO: this should be a 400 bad request
			return err
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
		// TODO: this should be a 400 bad request
		return fmt.Errorf("query parameter %q not set", name)
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 16)
		if err != nil {
			// TODO: this should be a 400 bad request
			return err
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
		// TODO: this should be a 400 bad request
		return fmt.Errorf("query parameter %q not set", name)
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 32)
		if err != nil {
			// TODO: this should be a 400 bad request
			return err
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
		// TODO: this should be a 400 bad request
		return fmt.Errorf("query parameter %q not set", name)
	}

	if o.required || (ok && val[0] != "") {
		intVal, err := strconv.ParseInt(val[0], 10, 64)
		if err != nil {
			// TODO: this should be a 400 bad request
			return err
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
