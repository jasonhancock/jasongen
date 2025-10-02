package validate

import (
	"fmt"
	"net/http"

	"github.com/jasonhancock/jasongen/internal/loader"
	"github.com/spf13/cobra"
)

// NewCmd sets up the command.
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "validate <file>",
		Short:        "Validates an openapi spec",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ex := executor{
				endpointValidationFuncs: []endpointValidationFunc{
					ruleRequire400,

					ruleRequireStatusCode(http.StatusInternalServerError),
					ruleRequireStatusCode(http.StatusBadGateway),
					ruleRequireStatusCode(http.StatusServiceUnavailable),
					ruleRequireStatusCode(http.StatusGatewayTimeout),

					ruleSecurityRequireStatusCode(http.StatusUnauthorized),
					ruleSecurityRequireStatusCode(http.StatusForbidden),

					ruleNoContentResponseBodyDefined,
					ruleRequireResponseBody(http.StatusOK),
					ruleRequireResponseBody(http.StatusCreated),
				},
			}

			errs := run(args[0], ex)
			if len(errs) > 0 {
				for _, v := range errs {
					fmt.Println(v)
				}
				return fmt.Errorf("%d validation errors encountered", len(errs))
			}
			return nil
		},
	}
}

type executor struct {
	endpointValidationFuncs []endpointValidationFunc
}

func run(file string, ex executor) []error {
	input, err := loader.MergeAndLoad(file)
	if err != nil {
		return []error{err}
	}

	var errs []error
	if input.Model.Paths != nil {
		for pair := input.Model.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
			path := pair.Key()
			pi := pair.Value()
			for opPair := pi.GetOperations().First(); opPair != nil; opPair = opPair.Next() {
				method := opPair.Key()
				op := opPair.Value()
				for _, fn := range ex.endpointValidationFuncs {
					if err := fn(method, path, op); err != nil {
						errs = append(errs, newEndpointError(method, path, err))
					}
				}
			}
		}
	}

	return errs
}
