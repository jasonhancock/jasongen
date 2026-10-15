package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/TwiN/deepmerge"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/utils"
)

func MergeFiles(files ...string) ([]byte, error) {
	if len(files) == 0 {
		return nil, errors.New("no files provided")
	}

	base, err := os.ReadFile(files[0])
	if err != nil {
		return nil, fmt.Errorf("loading base file: %w", err)
	}

	for i := 1; i < len(files); i++ {
		data, err := os.ReadFile(files[i])
		if err != nil {
			return nil, fmt.Errorf("loading file %q: %w", files[i], err)
		}

		base, err = deepmerge.YAML(base, data)
		if err != nil {
			return nil, fmt.Errorf("merging file %q: %w", files[i], err)
		}

	}

	return base, nil
}

func MergeAndLoad(files ...string) (*libopenapi.DocumentModel[v3high.Document], error) {
	data, err := MergeFiles(files...)
	if err != nil {
		return nil, err
	}

	return load(data)
}

func load(data []byte) (*libopenapi.DocumentModel[v3high.Document], error) {
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, err
	}

	if document.GetSpecInfo().SpecType == utils.OpenApi2 {
		document, err = convertSwagger2(document)
		if err != nil {
			return nil, fmt.Errorf("converting swagger 2.0 document to openapi 3: %w", err)
		}
	}

	// because we know this is a v3 spec, we can build a ready to go model from it.
	docModel, err := document.BuildV3Model()
	if err != nil {
		return nil, err
	}

	// Sort the tags
	sort.Slice(docModel.Model.Tags, func(i, j int) bool {
		return docModel.Model.Tags[i].Name < docModel.Model.Tags[j].Name
	})

	return docModel, nil
}

// convertSwagger2 converts a Swagger 2.0 document to OpenAPI 3 so the rest of the
// generator only has to deal with one spec version.
func convertSwagger2(document libopenapi.Document) (libopenapi.Document, error) {
	info := document.GetSpecInfo()
	data := info.GetSpecJSONBytes()
	if data == nil {
		if err := info.GetSpecJSONError(); err != nil {
			return nil, err
		}
		return nil, errors.New("document has no JSON representation")
	}

	var v2 openapi2.T
	if err := json.Unmarshal(*data, &v2); err != nil {
		return nil, fmt.Errorf("parsing swagger 2.0 document: %w", err)
	}

	v3, err := openapi2conv.ToV3(&v2)
	if err != nil {
		return nil, err
	}

	converted, err := json.Marshal(v3)
	if err != nil {
		return nil, err
	}

	return libopenapi.NewDocument(converted)
}
