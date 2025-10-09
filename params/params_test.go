package params

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	err := &missingParamErr{name: "some param"}
	require.Equal(t, http.StatusBadRequest, err.StatusCode())
}
