package merge

import (
	"bytes"
	"io"
	"os"

	"github.com/jasonhancock/jasongen/internal/loader"
	"github.com/spf13/cobra"
	"github.com/stuart-warren/yamlfmt"
)

// NewCmd sets up the command.
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "merge <base file> <file1> <file2> ... <fileN>",
		Short:        "Merges multiple OpenAPI definitions into a single spec",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(os.Stdout, args...)
		},
	}
}

func run(dest io.Writer, files ...string) error {
	model, err := loader.MergeAndLoad(files...)
	if err != nil {
		return err
	}
	result, err := model.Model.Render()
	if err != nil {
		return err
	}

	b, err := yamlfmt.Format(bytes.NewReader(result), true)
	if err != nil {
		return err
	}

	_, err = dest.Write(b)
	return err
}
