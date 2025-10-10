package params

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/jasonhancock/go-helpers"
	"github.com/stretchr/testify/require"
)

func TestQueryParamsString(t *testing.T) {
	tests := []struct {
		desc     string
		vals     url.Values
		expected *string
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			url.Values{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			url.Values{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			url.Values{},
			nil,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			url.Values{},
			nil,
			nil,
			[]Option{Required(false)},
		},
		{
			"enum, value set, required=true",
			url.Values{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			nil,
			[]Option{
				Required(true),
				EnumeratedValues(map[string]struct{}{
					"bar": {},
				}),
			},
		},
		{
			"enum, value set, required=false",
			url.Values{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			nil,
			[]Option{
				Required(false),
				EnumeratedValues(map[string]struct{}{
					"bar": {},
				}),
			},
		},
		{
			"enum, value set, required=true, bad enum",
			url.Values{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			errors.New(`"bar" is not a valid enumerated value`),
			[]Option{
				Required(true),
				EnumeratedValues(map[string]struct{}{
					"barbar": {},
				}),
			},
		},
		{
			"enum, value set, required=false, bad enum",
			url.Values{"foo": []string{"bar"}},
			helpers.Ptr("bar"),
			errors.New(`"bar" is not a valid enumerated value`),
			[]Option{
				Required(false),
				EnumeratedValues(map[string]struct{}{
					"barbar": {},
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			val, err := QueryParamString(tt.vals, "foo", tt.opts...)
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

func testGeneric[T any](t *testing.T, goodVal T, fmtStr string, fn func(url.Values, string, ...Option) (*T, error)) {
	t.Helper()
	tests := []struct {
		desc     string
		vals     url.Values
		expected *T
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			url.Values{"foo": []string{fmt.Sprintf(fmtStr, goodVal)}},
			&goodVal,
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			url.Values{"foo": []string{fmt.Sprintf(fmtStr, goodVal)}},
			&goodVal,
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			url.Values{},
			nil,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			url.Values{},
			nil,
			nil,
			[]Option{Required(false)},
		},
		{
			"value set, not an T",
			url.Values{"foo": []string{"foo"}},
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

func TestQueryParamsBool(t *testing.T) {
	testGeneric(t, bool(true), "%t", QueryParamBool)
}

func TestQueryParamsInt8(t *testing.T) {
	testGeneric(t, int8(64), "%d", QueryParamInt8)
}

func TestQueryParamsInt16(t *testing.T) {
	testGeneric(t, int16(1234), "%d", QueryParamInt16)
}
func TestQueryParamsInt32(t *testing.T) {
	testGeneric(t, int32(1234), "%d", QueryParamInt32)
}

func TestQueryParamsInt64(t *testing.T) {
	testGeneric(t, int64(1234), "%d", QueryParamInt64)
}

func TestQueryParamsFloat32(t *testing.T) {
	testGeneric(t, float32(1234.5), "%f", QueryParamFloat32)
}

func TestQueryParamsFloat64(t *testing.T) {
	testGeneric(t, 1234.5, "%f", QueryParamFloat64)
}
