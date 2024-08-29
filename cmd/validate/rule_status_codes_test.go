package validate

import (
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRules(t *testing.T) {
	tests := []struct {
		ruleName string
		rule     endpointValidationFunc
		input    string
		err      error
	}{
		{
			"ruleRequire400",
			ruleRequire400,
			"ok",
			nil,
		},
		{
			"ruleRequire400",
			ruleRequire400,
			"ok-get",
			nil,
		},
		{
			"ruleRequire400",
			ruleRequire400,
			"error",
			err400ResponseNotDefined,
		},

		{
			"ruleRequireStatusCode",
			ruleRequireStatusCode(http.StatusBadRequest),
			"ok",
			nil,
		},
		{
			"ruleRequireStatusCode",
			ruleRequireStatusCode(http.StatusBadRequest),
			"error",
			newStatusCodeMissingError(http.StatusBadRequest),
		},

		{
			"ruleRequire404IfParameterId",
			ruleRequire404IfParameterId,
			"ok",
			nil,
		},
		{
			"ruleRequire404IfParameterId",
			ruleRequire404IfParameterId,
			"ok-no-id",
			nil,
		},
		{
			"ruleRequire404IfParameterId",
			ruleRequire404IfParameterId,
			"error",
			newStatusCodeMissingError(http.StatusNotFound),
		},

		{
			"ruleNoContentResponseBodyDefined",
			ruleNoContentResponseBodyDefined,
			"ok",
			nil,
		},
		{
			"ruleNoContentResponseBodyDefined",
			ruleNoContentResponseBodyDefined,
			"error",
			err204ResponseBodyDefined,
		},

		{
			"ruleRequireResponseBody",
			ruleRequireResponseBody(http.StatusOK),
			"ok",
			nil,
		},
		{
			"ruleRequireResponseBody",
			ruleRequireResponseBody(http.StatusOK),
			"error",
			newResponseBodyRequired(http.StatusOK),
		},

		{
			"ruleSecurityRequireStatusCode",
			ruleSecurityRequireStatusCode(http.StatusForbidden),
			"ok",
			nil,
		},
		{
			"ruleSecurityRequireStatusCode",
			ruleSecurityRequireStatusCode(http.StatusForbidden),
			"error",
			newStatusCodeMissingError(http.StatusForbidden),
		},
	}

	for _, tt := range tests {
		t.Run(tt.ruleName+"/"+tt.input, func(t *testing.T) {
			ex := executor{
				endpointValidationFuncs: []endpointValidationFunc{tt.rule},
			}

			errs := run(filepath.Join("testdata", tt.ruleName, tt.input+".yaml"), ex)
			if tt.err == nil {
				require.Len(t, errs, 0)
				return
			}

			require.Len(t, errs, 1)
			require.Equal(t, tt.err.Error(), errors.Unwrap(errs[0]).Error())
		})
	}
}
