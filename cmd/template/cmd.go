package template

import (
	"embed"
	"fmt"
	"os"
	"path"

	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/jasongen/internal/loader"
	"github.com/spf13/cobra"
)

//go:embed templates
var templates embed.FS

// NewCmd sets up the command.
func NewCmd(info version.Info) *cobra.Command {
	var overwrite bool

	cmd := &cobra.Command{
		Use:          "template <package> <template> <outfile> <file1> <file2> ... <fileN>",
		Short:        "Renders a template",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				pkg     = args[0]
				tmpl    = args[1]
				outfile = args[2]
				files   = args[3:]
			)

			return runTemplate(pkg, tmpl, outfile, overwrite, info, files...)
		},
	}

	cmd.Flags().BoolVar(
		&overwrite,
		"overwrite",
		true,
		"When true, will blindly overwrite files.",
	)

	return cmd
}

func runTemplate(pkg, tmpl, outfile string, overwrite bool, info version.Info, files ...string) error {
	result, err := loader.MergeFiles(files...)
	if err != nil {
		return err
	}

	td, err := templateDataFrom(result, pkg, info)
	if err != nil {
		return err
	}

	var templateBytes string
	if data, err := templates.ReadFile(path.Join("templates", tmpl+".go.tmpl")); err == nil {
		templateBytes = string(data)
	} else {
		data, err := os.ReadFile(tmpl)
		if err != nil {
			return err
		}
		templateBytes = string(data)
	}

	{ // Determine if we need to write the original file or not.
		_, err = os.Stat(outfile)
		if err != nil && !os.IsNotExist(err) {
			// it was some other error
			return err
		}
		if err == nil && !overwrite {
			newOutfile := outfile + ".new"
			fmt.Fprintf(os.Stderr, "WARNING: output file (%q) exists. Writing output to %q instead\n", outfile, newOutfile)
			outfile = newOutfile
		}
	}

	fh, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer fh.Close()

	return renderTemplate(templateBytes, td, fh)
}
