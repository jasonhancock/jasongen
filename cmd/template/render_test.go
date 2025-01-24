package template

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerParameterizedURI(t *testing.T) {
	tests := []struct {
		input    string
		params   []Param
		expected string
		err      error
	}{
		{
			"/{id}",
			[]Param{
				{
					Name:     "id",
					Type:     "string",
					Location: "path",
				},
			},
			`fmt.Sprintf("/%s", id)`,
			nil,
		},
		{
			"/games/{game_id}",
			[]Param{
				{
					Name:     "game_id",
					Type:     "string",
					Location: "path",
				},
			},
			`fmt.Sprintf("/games/%s", gameID)`,
			nil,
		},
		{
			"/games/{game_id}/foo",
			[]Param{
				{
					Name:     "game_id",
					Type:     "string",
					Location: "path",
				},
			},
			`fmt.Sprintf("/games/%s/foo", gameID)`,
			nil,
		},
		{
			"/games/{game_id}/players/{player_id}",
			[]Param{
				{
					Name:     "game_id",
					Type:     "string",
					Location: "path",
				},
				{
					Name:     "player_id",
					Type:     "string",
					Location: "path",
				},
			},
			`fmt.Sprintf("/games/%s/players/%s", gameID, playerID)`,
			nil,
		},
		{
			// no path params
			"/games",
			nil,
			`"/games"`,
			nil,
		},
		{ // wildcard
			"/games/*",
			[]Param{
				{
					Name:          "wildcard",
					Type:          "string",
					Location:      "path",
					RetrievalName: "*",
				},
			},
			`fmt.Sprintf("/games/%s", wildcard)`,
			nil,
		},

		// TODO: add error cases (param not found in list)
		// TODO: add support for integer params
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			h := Handler{
				Path:   tt.input,
				Params: tt.params,
			}

			result, err := h.ParameterizedURI()
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldLess(t *testing.T) {
	tests := []struct {
		A        string
		B        string
		expected bool
	}{
		{"id", "aaa", true},
		{"created_at", "updated_at", true},
		{"created_at", "zzz", false},
		{"my_int", "created_at", true},
		{"updated_at", "zzz", false},
		{"foo", "bar", false},
		{"bar", "foo", true},
		{"groups", "id", false},
	}

	for _, tt := range tests {
		a := typeName(tt.A)
		b := typeName(tt.B)

		t.Run(fmt.Sprintf("%s %s", a, b), func(t *testing.T) {
			require.Equal(t, tt.expected, Field{Name: a}.Less(Field{Name: b}))
		})
	}
}

func TestTypeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"id", "ID"},
		{"api", "API"},
		{"http", "HTTP"},
		{"server_http_endpoint", "ServerHTTPEndpoint"},
		{"server_api", "ServerAPI"},
		{"foo", "Foo"},
		{"string", "string"},
		{"int32", "int32"},
		{"gid", "GID"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, typeName(tt.input))
		})
	}
}

func TestArgName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"id", "id"},
		{"api", "api"},
		{"http", "http"},
		{"server_http_endpoint", "serverHTTPEndpoint"},
		{"server_api", "serverAPI"},
		{"foo", "foo"},
		{"MyAuth", "myAuth"},
		{"groupName", "groupName"},
		{"type", "_type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, argName(tt.input))
		})
	}
}
