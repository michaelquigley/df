package df

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// BindFromJSON reads JSON from the specified file path and binds it to the target struct.
func BindFromJSON(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read JSON file %s: %w", path, err)
	}

	var jsonData map[string]any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON from %s: %w", path, err)
	}

	return Bind(target, jsonData, opts...)
}

// BindFromYAML reads YAML from the specified file path and binds it to the target struct.
func BindFromYAML(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read YAML file %s: %w", path, err)
	}

	var yamlData map[string]any
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to parse YAML from %s: %w", path, err)
	}

	return Bind(target, yamlData, opts...)
}

// UnbindToJSON converts a struct to map using Unbind, then writes it as JSON to the specified file path.
func UnbindToJSON(source interface{}, path string) error {
	data, err := Unbind(source)
	if err != nil {
		return fmt.Errorf("failed to unbind source: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", path, err)
	}

	return nil
}

// UnbindToYAML converts a struct to map using Unbind, then writes it as YAML to the specified file path.
func UnbindToYAML(source interface{}, path string) error {
	data, err := Unbind(source)
	if err != nil {
		return fmt.Errorf("failed to unbind source: %w", err)
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	if err := os.WriteFile(path, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file %s: %w", path, err)
	}

	return nil
}