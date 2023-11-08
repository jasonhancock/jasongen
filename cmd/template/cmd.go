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
		Use:          "template <base file> <file> <package> <template> <outfile>",
		Short:        "Renders a template",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				baseFile = args[0]
				file     = args[1]
				pkg      = args[2]
				tmpl     = args[3]
				outfile  = args[4]
			)

			result, err := loader.MergeFiles(baseFile, file)
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

			fh, err := os.Create(outfile)
			if err != nil {
				return err
			}
			defer fh.Close()

			return renderTemplate(templateBytes, td, fh)
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
