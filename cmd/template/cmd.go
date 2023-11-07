package template

import (
	"embed"
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

			fh, err := os.Create(outfile)
			if err != nil {
				return err
			}
			defer fh.Close()

			return renderTemplate(templateBytes, td, fh)
		},
	}

	return cmd
}
