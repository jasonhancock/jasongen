package loader

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/TwiN/deepmerge"
	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func MergeFiles(files ...string) (*libopenapi.DocumentModel[v3high.Document], error) {
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

	return load(base)
}

func load(data []byte) (*libopenapi.DocumentModel[v3high.Document], error) {
	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, err
	}

	// because we know this is a v3 spec, we can build a ready to go model from it.
	docModel, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	// Sort the tags
	sort.Slice(docModel.Model.Tags, func(i, j int) bool {
		return docModel.Model.Tags[i].Name < docModel.Model.Tags[j].Name
	})

	return docModel, nil
}
