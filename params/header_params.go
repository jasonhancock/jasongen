package params

import (
	"net/http"
)

func HeaderParamBool(values http.Header, name string, opts ...Option) (*bool, error) {
	return paramBool(map[string][]string(values), name, opts...)
}

func HeaderParamString(values http.Header, name string, opts ...Option) (*string, error) {
	return paramString(map[string][]string(values), name, opts...)
}

func HeaderParamInt8(values http.Header, name string, opts ...Option) (*int8, error) {
	return paramInt8(map[string][]string(values), name, opts...)
}

func HeaderParamInt16(values http.Header, name string, opts ...Option) (*int16, error) {
	return paramInt16(map[string][]string(values), name, opts...)
}

func HeaderParamInt32(values http.Header, name string, opts ...Option) (*int32, error) {
	return paramInt32(map[string][]string(values), name, opts...)
}

func HeaderParamInt64(values http.Header, name string, opts ...Option) (*int64, error) {
	return paramInt64(map[string][]string(values), name, opts...)
}

func HeaderParamFloat32(values http.Header, name string, opts ...Option) (*float32, error) {
	return paramFloat32(map[string][]string(values), name, opts...)
}

func HeaderParamFloat64(values http.Header, name string, opts ...Option) (*float64, error) {
	return paramFloat64(map[string][]string(values), name, opts...)
}
