package merge

import (
	"os"

	"github.com/jasonhancock/jasongen/internal/loader"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewCmd sets up the command.
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "merge <base file> <file1> <file2> ... <fileN>",
		Short:        "Merges multiple OpenAPI definitions into a single spec",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := loader.MergeFiles(args...)
			if err != nil {
				return err
			}

			return yaml.NewEncoder(os.Stdout).Encode(result.Model)
		},
	}
}
