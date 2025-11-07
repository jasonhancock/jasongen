package template

import (
	"embed"
	"fmt"
	"os"
	"path"

	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/cobraflags/root"
	"github.com/jasonhancock/jasongen/internal/loader"
	"github.com/spf13/cobra"
)

//go:embed templates
var templates embed.FS

type cmdOptions struct {
	overwrite bool
	pkgModels string
	language  string
}

// NewCmd sets up the command.
func NewCmd(r *root.Command) *cobra.Command {
	var opts cmdOptions

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

			return runTemplate(pkg, tmpl, outfile, opts, *r.Version, files...)
		},
	}

	cmd.Flags().BoolVar(
		&opts.overwrite,
		"overwrite",
		true,
		"When true, will blindly overwrite files.",
	)

	cmd.Flags().StringVar(
		&opts.pkgModels,
		"pkg-models",
		"",
		"The fully qualified import path to the models package.",
	)

	cmd.Flags().StringVar(
		&opts.language,
		"language",
		"go",
		"The language of the generated file (go|js).",
	)

	return cmd
}

func runTemplate(pkg, tmpl, outfile string, opts cmdOptions, info version.Info, files ...string) error {
	result, err := loader.MergeAndLoad(files...)
	if err != nil {
		return err
	}

	td, err := templateDataFrom(result, pkg, info, opts)
	if err != nil {
		return err
	}

	var templateBytes string
	if data, err := templates.ReadFile(path.Join("templates", tmpl+"."+opts.language+".tmpl")); err == nil {
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
		if err == nil && !opts.overwrite {
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

	return renderTemplate(templateBytes, td, fh, opts.pkgModels, opts.language)
}
