package validate

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jasonhancock/go-helpers"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

var err400ResponseNotDefined = newStatusCodeMissingError(http.StatusBadRequest)
var err204ResponseBodyDefined = errors.New("status 204 (No Content) but a response body is defined")

type endpointValidationFunc func(method, path string, op *v3high.Operation) error

// ruleRequire400 requires a 400 response to be defined for POST/PUT/PATCH operations.
// TODO: perhaps this should only matter if a content-type is specified, or if a request body is defined?
func ruleRequire400(method, path string, op *v3high.Operation) error {
	if !(method == "put" || method == "post" || method == "patch") {
		return nil
	}

	return ruleRequireStatusCode(http.StatusBadRequest)(method, path, op)
}

// This rule requires that a 400 response is defined if there is a path parameter that is an integer.
func ruleRequire400ParamInteger(method, path string, op *v3high.Operation) error {
	params := filterParametersType(filterParametersIn(op.Parameters, "path"), "integer")
	if len(params) == 0 {
		return nil
	}

	return ruleRequireStatusCode(http.StatusBadRequest)(method, path, op)
}

func filterParametersIn(input []*v3high.Parameter, in string) []*v3high.Parameter {
	params := make([]*v3high.Parameter, 0)
	for i := range input {
		if input[i].In == in {
			params = append(params, input[i])
		}
	}

	return params
}

func filterParametersType(input []*v3high.Parameter, t string) []*v3high.Parameter {
	params := make([]*v3high.Parameter, 0)
	for i := range input {
		if input[i].Schema != nil && helpers.Contains(input[i].Schema.Schema().Type, t) {
			params = append(params, input[i])
		}
	}

	return params
}

func ruleRequireStatusCode(code int) endpointValidationFunc {
	return func(method, path string, op *v3high.Operation) error {
		if op.Responses == nil {
			return newStatusCodeMissingError(code)
		}

		if _, ok := op.Responses.Codes.Get(strconv.Itoa(code)); !ok {
			return newStatusCodeMissingError(code)
		}

		return nil
	}
}

func ruleSecurityRequireStatusCode(code int) endpointValidationFunc {
	return func(method, path string, op *v3high.Operation) error {
		if len(op.Security) == 0 {
			return nil
		}

		return ruleRequireStatusCode(code)(method, path, op)
	}
}

func ruleRequire404IfParameterId(method, path string, op *v3high.Operation) error {
	params := parametersFromPath(path)
	hasId := false
	for _, p := range params {
		if p == "id" {
			hasId = true
			break
		}
	}

	if !hasId {
		return nil
	}

	if op.Responses == nil {
		return newStatusCodeMissingError(http.StatusNotFound)
	}

	return ruleRequireStatusCode(http.StatusNotFound)(method, path, op)
}

func ruleNoContentResponseBodyDefined(_, _ string, op *v3high.Operation) error {
	if op.Responses == nil {
		return nil
	}

	resp, ok := op.Responses.Codes.Get(strconv.Itoa(http.StatusNoContent))
	if !ok {
		return nil
	}

	if resp.Content != nil && resp.Content.Len() > 0 {
		return err204ResponseBodyDefined
	}

	return nil
}

func ruleRequireResponseBody(code int) endpointValidationFunc {
	return func(method, path string, op *v3high.Operation) error {
		if op.Responses == nil {
			return nil
		}

		resp, ok := op.Responses.Codes.Get(strconv.Itoa(code))
		if !ok {
			return nil
		}

		if resp.Content == nil || resp.Content.Len() == 0 {
			return newResponseBodyRequired(code)
		}

		return nil
	}
}
