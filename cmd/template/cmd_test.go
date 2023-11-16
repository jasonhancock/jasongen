package template

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

var flagSave = flag.Bool("save", false, "Will overwrite expected values files")

func TestRunTemplate(t *testing.T) {
	const baseFile = "testdata/openapi_base.yaml"
	const file = "testdata/openapi_sub.yaml"

	info := version.Info{
		Version: "1.2.3",
	}

	templates, err := filepath.Glob("templates/*.tmpl")
	require.NoError(t, err)

	dir := t.TempDir()
	for _, tmpl := range templates {
		tmpl = filepath.Base(tmpl)
		tmpl = strings.TrimSuffix(tmpl, ".go.tmpl")

		t.Run(tmpl, func(t *testing.T) {
			outfile := filepath.Join(dir, tmpl+".go")
			err := runTemplate("widgets", tmpl, outfile, false, info, baseFile, file)
			require.NoError(t, err)

			expectedFile := "testdata/expected/" + tmpl + ".txt"
			if *flagSave {
				b, err := os.ReadFile(outfile)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(expectedFile, b, 0644))
			}

			generic.FilesEqual(t, expectedFile, outfile)
		})
	}
}
