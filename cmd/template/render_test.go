package template

import (
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
			`fmt.Sprintf("/games/%s", gameId)`,
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
			`fmt.Sprintf("/games/%s/foo", gameId)`,
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
			`fmt.Sprintf("/games/%s/players/%s", gameId, playerId)`,
			nil,
		},
		{
			// no path params
			"/games",
			nil,
			`"/games"`,
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
