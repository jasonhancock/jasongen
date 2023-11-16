package template

import (
	"path/filepath"
	"strings"
	"testing"

	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

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

			generic.FilesEqual(t, "testdata/expected/"+tmpl+".txt", outfile)
		})
	}
}
