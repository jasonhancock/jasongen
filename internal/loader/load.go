package loader

import (
	"errors"
	"fmt"
	"os"

	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func MergeFiles(files ...string) (*libopenapi.DocumentModel[v3high.Document], error) {
	if len(files) == 0 {
		return nil, errors.New("no files provided")
	}
	base, err := loadFile(files[0])
	if err != nil {
		return nil, fmt.Errorf("loading base file: %w", err)
	}

	for i := 1; i < len(files); i++ {
		file, err := loadFile(files[i])
		if err != nil {
			return nil, fmt.Errorf("loading file %q: %w", files[i], err)
		}

		base, err = merge(base, file)
		if err != nil {
			return nil, fmt.Errorf("error when merging file %q: %w", files[i], err)
		}
	}

	return base, nil
}

func merge(base, file *libopenapi.DocumentModel[v3high.Document]) (*libopenapi.DocumentModel[v3high.Document], error) {
	// merge tags?

	// merge paths
	if file.Model.Paths != nil {
		if base.Model.Paths == nil {
			base.Model.Paths = file.Model.Paths
		} else {
			for pathName := range file.Model.Paths.PathItems {
				if _, ok := base.Model.Paths.PathItems[pathName]; ok {
					return nil, fmt.Errorf("duplicate path defined: %q", pathName)
				}

				base.Model.Paths.PathItems[pathName] = file.Model.Paths.PathItems[pathName]
			}
		}
	}

	// merge components
	if file.Model.Components != nil {
		if base.Model.Components == nil {
			base.Model.Components = file.Model.Components
		} else {
			// merge schemas
			if file.Model.Components.Schemas != nil {
				if base.Model.Components.Schemas == nil {
					base.Model.Components.Schemas = file.Model.Components.Schemas
				} else {

					for schemaName := range file.Model.Components.Schemas {
						if _, ok := base.Model.Components.Schemas[schemaName]; ok {
							return nil, fmt.Errorf("duplicate schema defined: %q", schemaName)
						}

						base.Model.Components.Schemas[schemaName] = file.Model.Components.Schemas[schemaName]
					}
				}
			}

			// merge the other shit too
		}
	}

	return base, nil
}

func loadFile(file string) (*libopenapi.DocumentModel[v3high.Document], error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	document, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, err
	}

	// because we know this is a v3 spec, we can build a ready to go model from it.
	docModel, errs := document.BuildV3Model()
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return docModel, nil
}
