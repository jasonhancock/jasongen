package merge

import (
	"bytes"
	"os"

	"github.com/jasonhancock/jasongen/internal/loader"
	"github.com/spf13/cobra"
	"github.com/stuart-warren/yamlfmt"
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

			var buf bytes.Buffer
			enc := yaml.NewEncoder(&buf)
			enc.SetIndent(2)
			if err := enc.Encode(result.Model); err != nil {
				return err
			}

			b, err := yamlfmt.Format(&buf, true)
			if err != nil {
				return err
			}

			_, err = os.Stdout.Write(b)
			return err
		},
	}
}
