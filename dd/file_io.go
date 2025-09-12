package dd

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v3"
)

// BindFromJSON reads JSON from the specified file path and binds it to the target struct.
func BindFromJSON(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read JSON", Cause: err}
	}

	var jsonData map[string]any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return &FileError{Path: path, Operation: "parse JSON from", Cause: err}
	}

	return Bind(target, jsonData, opts...)
}

// BindFromYAML reads YAML from the specified file path and binds it to the target struct.
func BindFromYAML(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read YAML", Cause: err}
	}

	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return &FileError{Path: path, Operation: "parse YAML from", Cause: err}
	}

	return Bind(target, yamlData, opts...)
}

// NewFromJSON reads JSON from the specified file path and returns a new instance of type T.
func NewFromJSON[T any](path string, opts ...*Options) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &FileError{Path: path, Operation: "read JSON", Cause: err}
	}

	var jsonData map[string]any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, &FileError{Path: path, Operation: "parse JSON from", Cause: err}
	}

	return New[T](jsonData, opts...)
}

// NewFromYAML reads YAML from the specified file path and returns a new instance of type T.
func NewFromYAML[T any](path string, opts ...*Options) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &FileError{Path: path, Operation: "read YAML", Cause: err}
	}

	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, &FileError{Path: path, Operation: "parse YAML from", Cause: err}
	}

	return New[T](yamlData, opts...)
}

// MergeFromJSON reads JSON from the specified file path and merges it with the target struct.
func MergeFromJSON(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read JSON", Cause: err}
	}

	var jsonData map[string]any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return &FileError{Path: path, Operation: "parse JSON from", Cause: err}
	}

	return Merge(target, jsonData, opts...)
}

// MergeFromYAML reads YAML from the specified file path and merges it with the target struct.
func MergeFromYAML(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read YAML", Cause: err}
	}

	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return &FileError{Path: path, Operation: "parse YAML from", Cause: err}
	}

	return Merge(target, yamlData, opts...)
}

// UnbindToJSON converts a struct to map using Unbind, then writes it as JSON to the specified file path.
func UnbindToJSON(source interface{}, path string) error {
	data, err := Unbind(source)
	if err != nil {
		return &ConversionError{Message: "failed to unbind source", Cause: err}
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return &ConversionError{Type: "JSON", Message: "failed to marshal", Cause: err}
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return &FileError{Path: path, Operation: "write JSON", Cause: err}
	}

	return nil
}

// UnbindToYAML converts a struct to map using Unbind, then writes it as YAML to the specified file path.
func UnbindToYAML(source interface{}, path string) error {
	data, err := Unbind(source)
	if err != nil {
		return &ConversionError{Message: "failed to unbind source", Cause: err}
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return &ConversionError{Type: "YAML", Message: "failed to marshal", Cause: err}
	}

	if err := os.WriteFile(path, yamlData, 0644); err != nil {
		return &FileError{Path: path, Operation: "write YAML", Cause: err}
	}

	return nil
}
