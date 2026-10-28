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

// externalModelsCases are the testdata/cases that are also rendered with the models
// in an external package, with the output compared against their expected_models dir.
var externalModelsCases = map[string]bool{
	"external_models": true,
}

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
						if tt.models != "" && !externalModelsCases[caseName] {
							// pointing at the non-existing github.com/example/somemodels package
							// takes forever, so only run it for the cases that opt in.
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
			"inline allOf property combining multiple schemas",
			`
components:
  schemas:
    Widget:
      type: object
      properties:
        owner:
          allOf:
            - $ref: '#/components/schemas/Owner'
            - $ref: '#/components/schemas/Name'
    Owner:
      type: object
      properties:
        id:
          type: string
    Name:
      type: object
      properties:
        name:
          type: string
`,
			"inline allOf schemas combining more than one schema are not supported",
		},
		{
			"security requirement needing two schemes at once",
			`
paths:
  /v1/widgets:
    delete:
      tags:
        - widgets
      operationId: WidgetsDelete
      security:
        - MyAuth: ['some_scope']
          different_auth: ['some_scope']
      responses:
        '204':
          description: successful operation
`,
			"requiring more than one security scheme at once is not supported yet",
		},
		{
			"empty security requirement among alternatives",
			`
paths:
  /v1/widgets:
    delete:
      tags:
        - widgets
      operationId: WidgetsDelete
      security:
        - MyAuth: ['some_scope']
        - {}
      responses:
        '204':
          description: successful operation
`,
			"an empty security requirement among alternatives is not supported yet",
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

// Swagger 2.0 documents are converted to OpenAPI 3 when loaded. This can't use the
// testdata/cases layout because those cases are merged with the OpenAPI 3 base file.
func TestRunTemplateSwagger2(t *testing.T) {
	for _, tmpl := range []string{"client", "models"} {
		t.Run(tmpl, func(t *testing.T) {
			outfile := filepath.Join(t.TempDir(), tmpl+".go")
			err := runTemplate(
				"widgets",
				tmpl,
				outfile,
				cmdOptions{language: "go"},
				version.Info{Version: "1.2.3"},
				filepath.Join("testdata", "swagger2", "swagger.yaml"),
			)
			require.NoError(t, err)

			expectedFile := filepath.Join("testdata", "swagger2", "expected", tmpl+".go.txt")
			if *flagSave {
				b, err := os.ReadFile(outfile)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(expectedFile, b, 0644))
			}

			generic.FilesEqual(t, expectedFile, outfile)
		})
	}
}
