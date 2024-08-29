package validate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParametersFromPath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"/", []string{}},
		{"/foo", []string{}},
		{"/foo/{id}", []string{"id"}},
		{"/foo/{id}/blah/{bar}", []string{"id", "bar"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expected, parametersFromPath(tt.input))
		})
	}
}
