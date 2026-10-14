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

func QueryParamUint8(values url.Values, name string, opts ...Option) (*uint8, error) {
	return paramUint8(map[string][]string(values), name, opts...)
}

func QueryParamUint16(values url.Values, name string, opts ...Option) (*uint16, error) {
	return paramUint16(map[string][]string(values), name, opts...)
}

func QueryParamUint32(values url.Values, name string, opts ...Option) (*uint32, error) {
	return paramUint32(map[string][]string(values), name, opts...)
}

func QueryParamUint64(values url.Values, name string, opts ...Option) (*uint64, error) {
	return paramUint64(map[string][]string(values), name, opts...)
}

func QueryParamBoolSlice(values url.Values, name string, opts ...Option) ([]bool, error) {
	return paramSlice(values, name, QueryParamBool, opts...)
}

func QueryParamStringSlice(values url.Values, name string, opts ...Option) ([]string, error) {
	return paramSlice(values, name, QueryParamString, opts...)
}

func QueryParamInt8Slice(values url.Values, name string, opts ...Option) ([]int8, error) {
	return paramSlice(values, name, QueryParamInt8, opts...)
}

func QueryParamInt16Slice(values url.Values, name string, opts ...Option) ([]int16, error) {
	return paramSlice(values, name, QueryParamInt16, opts...)
}

func QueryParamInt32Slice(values url.Values, name string, opts ...Option) ([]int32, error) {
	return paramSlice(values, name, QueryParamInt32, opts...)
}

func QueryParamInt64Slice(values url.Values, name string, opts ...Option) ([]int64, error) {
	return paramSlice(values, name, QueryParamInt64, opts...)
}

func QueryParamFloat32Slice(values url.Values, name string, opts ...Option) ([]float32, error) {
	return paramSlice(values, name, QueryParamFloat32, opts...)
}

func QueryParamFloat64Slice(values url.Values, name string, opts ...Option) ([]float64, error) {
	return paramSlice(values, name, QueryParamFloat64, opts...)
}

func QueryParamUint8Slice(values url.Values, name string, opts ...Option) ([]uint8, error) {
	return paramSlice(values, name, QueryParamUint8, opts...)
}

func QueryParamUint16Slice(values url.Values, name string, opts ...Option) ([]uint16, error) {
	return paramSlice(values, name, QueryParamUint16, opts...)
}

func QueryParamUint32Slice(values url.Values, name string, opts ...Option) ([]uint32, error) {
	return paramSlice(values, name, QueryParamUint32, opts...)
}

func QueryParamUint64Slice(values url.Values, name string, opts ...Option) ([]uint64, error) {
	return paramSlice(values, name, QueryParamUint64, opts...)
}
