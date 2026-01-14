package da

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/michaelquigley/df/dd"
)

// Loader defines how configuration is loaded from a source.
type Loader interface {
	Load(dest any) error
}

// fileLoader implements Loader for JSON/YAML files.
type fileLoader struct {
	paths    []string
	optional bool
}

// FileLoader creates a loader for required config files.
// File format is determined by extension (.json, .yaml, .yml).
// Returns error if any file doesn't exist or can't be parsed.
func FileLoader(paths ...string) Loader {
	return &fileLoader{paths: paths, optional: false}
}

// OptionalFileLoader creates a loader that skips missing files.
// File format is determined by extension (.json, .yaml, .yml).
// Missing files are silently skipped; other errors are returned.
func OptionalFileLoader(paths ...string) Loader {
	return &fileLoader{paths: paths, optional: true}
}

func (l *fileLoader) Load(dest any) error {
	for _, path := range l.paths {
		ext := filepath.Ext(path)
		var err error
		switch ext {
		case ".yaml", ".yml":
			err = dd.MergeYAMLFile(dest, path)
		case ".json":
			err = dd.MergeJSONFile(dest, path)
		default:
			return fmt.Errorf("unsupported config extension: %s", ext)
		}
		if err != nil {
			// check if it's a not-found error and optional
			var fileErr *dd.FileError
			if l.optional && errors.As(err, &fileErr) && fileErr.IsNotFound() {
				continue
			}
			return err
		}
	}
	return nil
}

// chainLoader combines multiple loaders.
type chainLoader struct {
	loaders []Loader
}

// ChainLoader creates a loader that applies multiple loaders in sequence.
// Returns on first error.
func ChainLoader(loaders ...Loader) Loader {
	return &chainLoader{loaders: loaders}
}

func (c *chainLoader) Load(dest any) error {
	for _, l := range c.loaders {
		if err := l.Load(dest); err != nil {
			return err
		}
	}
	return nil
}

// Config populates a config struct using the provided loaders.
// Loaders are applied in sequence; returns on first error.
func Config[C any](cfg *C, loaders ...Loader) error {
	for _, l := range loaders {
		if err := l.Load(cfg); err != nil {
			return err
		}
	}
	return nil
}
