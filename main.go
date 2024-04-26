package main

import (
	ver "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/cobraflags/root"
	"github.com/jasonhancock/jasongen/cmd/merge"
	"github.com/jasonhancock/jasongen/cmd/template"
)

// These variables are populated by goreleaser when the binary is built.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	info := ver.New(version, commit, date)

	r := root.New(
		"jasongen",
		root.WithShort("Jason's OpenAPI Code Generator"),
		root.WithVersion(info),
		root.LoggerEnabled(true),
		root.WithCommand(
			merge.NewCmd(),
			ver.NewCmd(*info),
		),
	)

	r.AddCommand(
		template.NewCmd(r),
	)

	r.Execute()
}
