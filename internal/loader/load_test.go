package loader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestMergeAndLoadSwagger2(t *testing.T) {
	yamlFile := filepath.Join("testdata", "swagger2.yaml")

	// The same document as JSON, to cover both input formats.
	b, err := os.ReadFile(yamlFile)
	require.NoError(t, err)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(b, &doc))
	jb, err := json.Marshal(doc)
	require.NoError(t, err)
	jsonFile := filepath.Join(t.TempDir(), "swagger2.json")
	require.NoError(t, os.WriteFile(jsonFile, jb, 0644))

	for _, file := range []string{yamlFile, jsonFile} {
		t.Run(filepath.Ext(file), func(t *testing.T) {
			model, err := MergeAndLoad(file)
			require.NoError(t, err)

			require.Contains(t, model.Model.Version, "3.")

			widget, ok := model.Model.Components.Schemas.Get("Widget")
			require.True(t, ok, "definitions should be converted to components/schemas")
			require.Equal(t, []string{"name"}, widget.Schema().Required)

			pi, ok := model.Model.Paths.PathItems.Get("/widgets/{id}")
			require.True(t, ok)

			// body parameters become request bodies, using the document's consumes.
			require.NotNil(t, pi.Put.RequestBody)
			body, ok := pi.Put.RequestBody.Content.Get("application/json")
			require.True(t, ok)
			require.Equal(t, "#/components/schemas/Widget", body.Schema.GetReference())

			resp, ok := pi.Get.Responses.Codes.Get("200")
			require.True(t, ok)
			content, ok := resp.Content.Get("application/json")
			require.True(t, ok)
			require.Equal(t, "#/components/schemas/Widget", content.Schema.GetReference())

			require.Len(t, pi.Get.Parameters, 1)
			require.Equal(t, "path", pi.Get.Parameters[0].In)
			require.Equal(t, []string{"integer"}, pi.Get.Parameters[0].Schema.Schema().Type)
		})
	}
}

func TestMergeAndLoadOpenAPI3(t *testing.T) {
	model, err := MergeAndLoad(filepath.Join("..", "..", "cmd", "template", "testdata", "openapi_base.yaml"))
	require.NoError(t, err)
	require.Equal(t, "3.1.0", model.Model.Version)
}
