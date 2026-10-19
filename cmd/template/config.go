package template

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/kenshaw/snaker"
	"go.yaml.in/yaml/v4"
)

// config holds generator settings that don't belong in the spec, loaded from the file
// passed via --config. Keeping them out of the spec means they work for specs you
// can't (or don't want to) edit.
type config struct {
	// Initialisms are added to the default set used when building Go identifiers,
	// so "ipam_prefix" becomes "IPAMPrefix" instead of "IpamPrefix".
	Initialisms []string `yaml:"initialisms"`
}

var validInitialism = regexp.MustCompile(`^[A-Za-z0-9]{2,}$`)

func loadConfig(file string) (config, error) {
	var cfg config
	if file == "" {
		return cfg, nil
	}

	fh, err := os.Open(file)
	if err != nil {
		return cfg, fmt.Errorf("opening config file: %w", err)
	}
	defer fh.Close()

	dec := yaml.NewDecoder(fh)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
		return cfg, fmt.Errorf("parsing config file %q: %w", file, err)
	}

	for _, v := range cfg.Initialisms {
		if !validInitialism.MatchString(v) {
			return cfg, fmt.Errorf("config file %q: invalid initialism %q: must be at least 2 letters or digits", file, v)
		}
	}

	return cfg, nil
}

// initialisms is the set used by the naming helpers. runTemplate replaces it for the
// duration of a run when the config adds initialisms.
var initialisms = mustInitialisms()

func newInitialisms(extra ...string) (*snaker.Initialisms, error) {
	ini, err := snaker.NewDefaultInitialisms()
	if err != nil {
		return nil, err
	}
	if err := ini.Add(extra...); err != nil {
		return nil, err
	}
	// Register plurals the same way snaker does for its defaults, so "RTTs" stays
	// "RTTs" instead of becoming "RTTS".
	pairs := make([]string, 0, len(extra)*2)
	for _, v := range extra {
		v = strings.ToUpper(v)
		pairs = append(pairs, v+"S", v+"s")
	}
	if err := ini.Post(pairs...); err != nil {
		return nil, err
	}
	return ini, nil
}

func mustInitialisms() *snaker.Initialisms {
	ini, err := newInitialisms()
	if err != nil {
		panic(fmt.Errorf("building initialisms: %w", err))
	}
	return ini
}
