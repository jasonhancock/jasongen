package template

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	version "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/go-logger"
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

	l := logger.Default()

	for _, tmpl := range templates {
		tmpl = filepath.Base(tmpl)
		tmpl = strings.TrimSuffix(tmpl, ".go.tmpl")

		t.Run(tmpl, func(t *testing.T) {
			tests := []struct {
				models      string
				expectedDir string
			}{
				{"", "expected"},
				{"github.com/example/somemodels", "expected_models"},
			}

			for _, tt := range tests {
				dir := t.TempDir()
				t.Run("models("+tt.models+")", func(t *testing.T) {
					opts := cmdOptions{
						overwrite: false,
						pkgModels: tt.models,
					}
					outfile := filepath.Join(dir, tmpl+".go")
					err := runTemplate(l, "widgets", tmpl, outfile, opts, info, baseFile, file)
					if err != nil {
						if _, err := os.Stat(outfile); err == nil {
							tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(outfile)))
							generic.CopyFile(t, outfile, tmpOut)
							t.Logf("output written to file %s", tmpOut)
						}
					}
					require.NoError(t, err)

					expectedFile := filepath.Join("testdata", tt.expectedDir, tmpl+".txt")
					if *flagSave {
						b, err := os.ReadFile(outfile)
						require.NoError(t, err)
						require.NoError(t, os.WriteFile(expectedFile, b, 0644))
					}

					generic.FilesEqual(t, expectedFile, outfile)
				})
			}
		})
	}
}
