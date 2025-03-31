package params

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryParamsString(t *testing.T) {
	tests := []struct {
		desc     string
		vals     url.Values
		expected string
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			url.Values{"foo": []string{"bar"}},
			"bar",
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			url.Values{"foo": []string{"bar"}},
			"bar",
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			url.Values{},
			"",
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			url.Values{},
			"",
			nil,
			[]Option{Required(false)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var dest string
			err := QueryParamString(tt.vals, "foo", &dest, tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, dest)
		})
	}
}

func TestQueryParamsInt32(t *testing.T) {
	tests := []struct {
		desc     string
		vals     url.Values
		expected int32
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			url.Values{"foo": []string{"1234"}},
			1234,
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			url.Values{"foo": []string{"1234"}},
			1234,
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			url.Values{},
			0,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			url.Values{},
			0,
			nil,
			[]Option{Required(false)},
		},
		// TODO: tests that parse the value not as an int
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var dest int32
			err := QueryParamInt32(tt.vals, "foo", &dest, tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, dest)
		})
	}
}

func TestQueryParamsInt64(t *testing.T) {
	tests := []struct {
		desc     string
		vals     url.Values
		expected int64
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			url.Values{"foo": []string{"1234"}},
			1234,
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			url.Values{"foo": []string{"1234"}},
			1234,
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			url.Values{},
			0,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			url.Values{},
			0,
			nil,
			[]Option{Required(false)},
		},
		// TODO: tests that parse the value not as an int
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var dest int64
			err := QueryParamInt64(tt.vals, "foo", &dest, tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, dest)
		})
	}
}

func TestQueryParamsBool(t *testing.T) {
	tests := []struct {
		desc     string
		vals     url.Values
		expected bool
		err      error
		opts     []Option
	}{
		{
			"value set, required=true",
			url.Values{"foo": []string{"true"}},
			true,
			nil,
			[]Option{Required(true)},
		},
		{
			"value set, required=false",
			url.Values{"foo": []string{"true"}},
			true,
			nil,
			[]Option{Required(false)},
		},
		{
			"value not set, required=true",
			url.Values{},
			false,
			errors.New(`query parameter "foo" not set`),
			[]Option{Required(true)},
		},
		{
			"value not set, required=false",
			url.Values{},
			false,
			nil,
			[]Option{Required(false)},
		},
		{
			"value not a bool, required=true",
			url.Values{"foo": []string{"blah"}},
			false,
			errors.New("invalid syntax"),
			[]Option{Required(true)},
		},
		{
			"value not a bool, required=false",
			url.Values{"foo": []string{"blah"}},
			false,
			errors.New("invalid syntax"),
			[]Option{Required(false)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var dest bool
			err := QueryParamBool(tt.vals, "foo", &dest, tt.opts...)
			if tt.err != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.err.Error())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, dest)
		})
	}
}
