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

				parts := strings.Split(tmpl, ".")

				require.Len(t, parts, 3)

				tmpl = parts[0]
				lang := parts[1]

				t.Run(fmt.Sprintf("%s.%s", tmpl, lang), func(t *testing.T) {
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
								language:  lang,
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

							expectedFile := filepath.Join("testdata", "cases", caseName, tt.expectedDir, fmt.Sprintf("%s.%s.txt", tmpl, lang))
							if *flagSave {
								b, err := os.ReadFile(outfile)
								require.NoError(t, err)
								require.NoError(t, os.MkdirAll(filepath.Dir(expectedFile), 0755))
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

func TestRunTemplateErrors(t *testing.T) {
	tests := []struct {
		desc     string
		spec     string
		contains string
	}{
		{
			"inline object property",
			`
paths:
  /v1/widgets:
    get:
      tags:
        - widgets
      operationId: WidgetGet
      responses:
        '200':
          description: successful operation
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Widget'
components:
  schemas:
    Widget:
      type: object
      properties:
        owner:
          type: object
          properties:
            name:
              type: string
`,
			"inline object schemas are only supported as request bodies",
		},
		{
			"inline object response",
			`
paths:
  /v1/widgets:
    get:
      tags:
        - widgets
      operationId: WidgetGet
      responses:
        '200':
          description: successful operation
          content:
            application/json:
              schema:
                type: object
                properties:
                  name:
                    type: string
`,
			"inline object schemas are only supported as request bodies",
		},
		{
			"request body name collides with a schema",
			`
paths:
  /v1/widgets:
    post:
      tags:
        - widgets
      operationId: WidgetCreate
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
      responses:
        '204':
          description: successful operation
components:
  schemas:
    WidgetCreateRequest:
      type: object
      properties:
        name:
          type: string
`,
			`multiple models named "WidgetCreateRequest"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			dir := t.TempDir()
			spec := filepath.Join(dir, "openapi.yaml")
			require.NoError(t, os.WriteFile(spec, []byte(tt.spec), 0644))

			err := runTemplate(
				"widgets",
				"models",
				filepath.Join(dir, "models.go"),
				cmdOptions{language: "go"},
				version.Info{Version: "1.2.3"},
				"testdata/openapi_base.yaml",
				spec,
			)
			require.ErrorContains(t, err, tt.contains)
		})
	}
}
