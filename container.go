package df

import (
	"errors"
	"fmt"
)

// Container provides an application container that manages configuration
// by starting with code-based defaults and overlaying external config files.
type Container[C any] struct {
	cfg C
}

// NewContainer creates a new container with the provided default configuration.
// The cfg parameter should contain code-based defaults that will be used as the base.
func NewContainer[C any](cfg C) *Container[C] {
	return &Container[C]{cfg: cfg}
}

// Config returns the current configuration instance.
func (c *Container[C]) Config() *C {
	return &c.cfg
}

// MergeFromJSON reads JSON configuration from the specified file path and merges
// it with the existing configuration. The merge operation overlays the file data
// onto the current configuration, preserving existing values for fields not
// present in the file.
func (c *Container[C]) MergeFromJSON(path string, opts ...*Options) error {
	if err := MergeFromJSON(&c.cfg, path, opts...); err != nil {
		return fmt.Errorf("failed to merge config from JSON file '%s': %w", path, err)
	}
	return nil
}

// MergeFromYAML reads YAML configuration from the specified file path and merges
// it with the existing configuration. The merge operation overlays the file data
// onto the current configuration, preserving existing values for fields not
// present in the file.
func (c *Container[C]) MergeFromYAML(path string, opts ...*Options) error {
	if err := MergeFromYAML(&c.cfg, path, opts...); err != nil {
		return fmt.Errorf("failed to merge config from YAML file '%s': %w", path, err)
	}
	return nil
}

// LoadConfigFiles merges configuration from multiple files in the order provided.
// JSON files are identified by .json extension, YAML files by .yaml or .yml extensions.
// If a file doesn't exist, it's skipped unless the Options specify otherwise.
func (c *Container[C]) LoadConfigFiles(paths []string, opts ...*Options) error {
	for _, path := range paths {
		var err error
		if isJSONFile(path) {
			err = c.MergeFromJSON(path, opts...)
		} else if isYAMLFile(path) {
			err = c.MergeFromYAML(path, opts...)
		} else {
			return fmt.Errorf("unsupported config file type: %s", path)
		}

		if err != nil {
			// check if it's a file not found error and skip if desired
			var fileErr *FileError
			if errors.As(err, &fileErr) && fileErr.IsNotFound() {
				continue
			}
			return err
		}
	}
	return nil
}

// isJSONFile checks if the file path has a JSON extension.
func isJSONFile(path string) bool {
	return len(path) >= 5 && path[len(path)-5:] == ".json"
}

// isYAMLFile checks if the file path has a YAML extension.
func isYAMLFile(path string) bool {
	return (len(path) >= 5 && path[len(path)-5:] == ".yaml") ||
		(len(path) >= 4 && path[len(path)-4:] == ".yml")
}
