package params

import "net/url"

func QueryParamBool(values url.Values, name string, opts ...Option) (*bool, error) {
	return paramBool(map[string][]string(values), name, opts...)
}

func QueryParamString(values url.Values, name string, opts ...Option) (*string, error) {
	return paramString(map[string][]string(values), name, opts...)
}

func QueryParamInt8(values url.Values, name string, opts ...Option) (*int8, error) {
	return paramInt8(map[string][]string(values), name, opts...)
}

func QueryParamInt16(values url.Values, name string, opts ...Option) (*int16, error) {
	return paramInt16(map[string][]string(values), name, opts...)
}

func QueryParamInt32(values url.Values, name string, opts ...Option) (*int32, error) {
	return paramInt32(map[string][]string(values), name, opts...)
}

func QueryParamInt64(values url.Values, name string, opts ...Option) (*int64, error) {
	return paramInt64(map[string][]string(values), name, opts...)
}

func QueryParamFloat32(values url.Values, name string, opts ...Option) (*float32, error) {
	return paramFloat32(map[string][]string(values), name, opts...)
}

func QueryParamFloat64(values url.Values, name string, opts ...Option) (*float64, error) {
	return paramFloat64(map[string][]string(values), name, opts...)
}
