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
	"github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

var flagSave = flag.Bool("save", false, "Will overwrite expected values files")

func TestRunTemplate(t *testing.T) {
	cases, err := filepath.Glob("testdata/cases/*")
	require.NoError(t, err)

	templates, err := filepath.Glob("templates/*.tmpl")
	require.NoError(t, err)

	for _, caseName := range cases {
		caseName = filepath.Base(caseName)
		file := filepath.Join("testdata", "cases", caseName, "openapi.yaml")

		t.Run(caseName, func(t *testing.T) {
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
						if caseName != "all" && tt.models != "" {
							// pointing at the non-existing github.com/example/somemodels package
							// takes forever, so just run it against the "all" test suite.
							continue
						}
						dir := t.TempDir()
						t.Run("models("+tt.models+")", func(t *testing.T) {
							opts := cmdOptions{
								overwrite: false,
								pkgModels: tt.models,
							}
							outfile := filepath.Join(dir, tmpl+".go")
							err := runTemplate(
								"widgets",
								tmpl,
								outfile,
								opts,
								version.Info{Version: "1.2.3"},
								"testdata/openapi_base.yaml",
								file,
							)
							if err != nil {
								if _, err := os.Stat(outfile); err == nil {
									tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(outfile)))
									generic.CopyFile(t, outfile, tmpOut)
									t.Logf("output written to file %s", tmpOut)
								}
							}
							require.NoError(t, err)

							expectedFile := filepath.Join("testdata", "cases", caseName, tt.expectedDir, tmpl+".txt")
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
		})
	}
}
