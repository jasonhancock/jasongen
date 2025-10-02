package merge

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

var flagSave = flag.Bool("save", false, "Will overwrite expected values files")

func TestCmd(t *testing.T) {
	files := []string{
		"testdata/openapi_base.yaml",
		"testdata/openapi_sub1.yaml",
		"testdata/openapi_sub2.yaml",
	}

	expectedFile := "testdata/expected.yaml"

	dir := t.TempDir()
	outfile := filepath.Join(dir, "out.yaml")
	fh, err := os.Create(outfile)
	require.NoError(t, err)
	defer fh.Close()

	require.NoError(t, run(fh, files...))
	require.NoError(t, fh.Close())

	if *flagSave {
		b, err := os.ReadFile(outfile)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(expectedFile, b, 0644))
	}

	generic.FilesEqual(t, expectedFile, outfile)
}
