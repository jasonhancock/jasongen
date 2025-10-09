package params

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jasonhancock/go-helpers"
	"github.com/stretchr/testify/require"
)

func TestHeaderParamsString(t *testing.T) {
	tests := []struct {
		desc     string
		vals     http.Header
		expected *string
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			http.Header{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			http.Header{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			http.Header{},
			nil,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			http.Header{},
			nil,
			nil,
			[]Option{Required(false)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			val, err := HeaderParamString(tt.vals, "foo", tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)

			if tt.expected == nil {
				require.Nil(t, val)
				return
			}
			require.NotNil(t, val)
			require.Equal(t, *tt.expected, *val)
		})
	}
}

func testGenericHeader[T any](t *testing.T, goodVal T, fmtStr string, fn func(http.Header, string, ...Option) (*T, error)) {
	t.Helper()
	tests := []struct {
		desc     string
		vals     http.Header
		expected *T
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			http.Header{"foo": []string{fmt.Sprintf(fmtStr, goodVal)}},
			&goodVal,
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			http.Header{"foo": []string{fmt.Sprintf(fmtStr, goodVal)}},
			&goodVal,
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			http.Header{},
			nil,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			http.Header{},
			nil,
			nil,
			[]Option{Required(false)},
		},
		{
			"value set, not an T",
			http.Header{"foo": []string{"foo"}},
			nil,
			errors.New("invalid syntax"),
			[]Option{Required(false)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			val, err := fn(tt.vals, "foo", tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)

			if tt.expected == nil {
				require.Nil(t, val)
				return
			}
			require.NotNil(t, val)
			require.Equal(t, *tt.expected, *val)
		})
	}
}

func TestHeaderParamsBool(t *testing.T) {
	testGenericHeader(t, bool(true), "%t", HeaderParamBool)
}

func TestHeaderParamsInt8(t *testing.T) {
	testGenericHeader(t, int8(64), "%d", HeaderParamInt8)
}

func TestHeaderParamsInt16(t *testing.T) {
	testGenericHeader(t, int16(1234), "%d", HeaderParamInt16)
}
func TestHeaderParamsInt32(t *testing.T) {
	testGenericHeader(t, int32(1234), "%d", HeaderParamInt32)
}

func TestHeaderParamsInt64(t *testing.T) {
	testGenericHeader(t, int64(1234), "%d", HeaderParamInt64)
}

func TestHeaderParamsFloat32(t *testing.T) {
	testGenericHeader(t, float32(1234.5), "%f", HeaderParamFloat32)
}

func TestHeaderParamsFloat64(t *testing.T) {
	testGenericHeader(t, 1234.5, "%f", HeaderParamFloat64)
}
