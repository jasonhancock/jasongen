package main

import (
	"log"
	"os"

	"github.com/jasonhancock/jasongen/internal/loader"
	"gopkg.in/yaml.v3"
)

func main() {
	baseFile := "/Users/jhancock/development/polaris-usersync/pkg/projects/openapi_base.yaml"

	files := []string{
		"/Users/jhancock/development/polaris-usersync/pkg/projects/openapi.yaml",
		//"/Users/jhancock/development/polaris-usersync/pkg/projects/openapi2.yaml",
	}

	result, err := loader.MergeFiles(baseFile, files...)
	if err != nil {
		log.Fatal(err)
	}

	yaml.NewEncoder(os.Stdout).Encode(result.Model)
}
