package template

import (
	"os"
	"path/filepath"
	"testing"

	version "github.com/jasonhancock/cobra-version"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		desc     string
		contents string
		expected config
		errMsg   string
	}{
		{"empty file", "", config{}, ""},
		{"initialisms", "initialisms:\n  - IPAM\n  - vlan\n", config{Initialisms: []string{"IPAM", "vlan"}}, ""},
		{"unknown key", "initialism:\n  - IPAM\n", config{}, "field initialism not found"},
		{"too short", "initialisms:\n  - X\n", config{}, `invalid initialism "X"`},
		{"invalid characters", "initialisms:\n  - IP-AM\n", config{}, `invalid initialism "IP-AM"`},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "jasongen.yaml")
			require.NoError(t, os.WriteFile(file, []byte(tt.contents), 0644))

			cfg, err := loadConfig(file)
			if tt.errMsg != "" {
				require.ErrorContains(t, err, tt.errMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, cfg)
		})
	}

	t.Run("no file", func(t *testing.T) {
		cfg, err := loadConfig("")
		require.NoError(t, err)
		require.Equal(t, config{}, cfg)
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadConfig(filepath.Join(t.TempDir(), "nope.yaml"))
		require.ErrorContains(t, err, "opening config file")
	})
}

func TestRunTemplateConfigInitialisms(t *testing.T) {
	const spec = `
paths:
  /v1/ipam/prefix:
    get:
      tags:
        - widgets
      operationId: ipam_prefixes_get
      responses:
        '200':
          description: successful operation
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ipam_prefix'
components:
  schemas:
    ipam_prefix:
      type: object
      properties:
        vlan_id:
          type: integer
        vlans:
          type: array
          items:
            type: string
`
	dir := t.TempDir()
	specFile := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specFile, []byte(spec), 0644))
	cfgFile := filepath.Join(dir, "jasongen.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("initialisms:\n  - IPAM\n  - VLAN\n"), 0644))

	render := func(t *testing.T, tmpl, configFile string) string {
		t.Helper()
		outfile := filepath.Join(t.TempDir(), tmpl+".go")
		require.NoError(t, runTemplate(
			"widgets",
			tmpl,
			outfile,
			cmdOptions{language: "go", configFile: configFile},
			version.Info{Version: "1.2.3"},
			"testdata/openapi_base.yaml",
			specFile,
		))
		b, err := os.ReadFile(outfile)
		require.NoError(t, err)
		return string(b)
	}

	t.Run("with config", func(t *testing.T) {
		models := render(t, "models", cfgFile)
		require.Contains(t, models, "type IPAMPrefix struct")
		require.Contains(t, models, "VLANID *int64")
		require.Contains(t, models, "VLANs  []string")

		client := render(t, "client", cfgFile)
		require.Contains(t, client, "func (c *Client) IPAMPrefixesGet(ctx context.Context) (IPAMPrefix, error)")
	})

	// Runs after the config run to make sure its initialisms don't leak into later runs.
	t.Run("without config", func(t *testing.T) {
		models := render(t, "models", "")
		require.Contains(t, models, "type IpamPrefix struct")
		require.Contains(t, models, "VlanID *int64")
	})
}
