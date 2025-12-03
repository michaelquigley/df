package dd

import (
	"encoding/json"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// --- Bytes Layer (base) ---

// BindJSON parses JSON data and binds it to the target struct.
func BindJSON(target interface{}, data []byte, opts ...*Options) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return &ConversionError{Type: "JSON", Message: "failed to parse", Cause: err}
	}
	return Bind(target, m, opts...)
}

// BindYAML parses YAML data and binds it to the target struct.
func BindYAML(target interface{}, data []byte, opts ...*Options) error {
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return &ConversionError{Type: "YAML", Message: "failed to parse", Cause: err}
	}
	return Bind(target, m, opts...)
}

// NewJSON parses JSON data and returns a new instance of type T.
func NewJSON[T any](data []byte, opts ...*Options) (*T, error) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, &ConversionError{Type: "JSON", Message: "failed to parse", Cause: err}
	}
	return New[T](m, opts...)
}

// NewYAML parses YAML data and returns a new instance of type T.
func NewYAML[T any](data []byte, opts ...*Options) (*T, error) {
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, &ConversionError{Type: "YAML", Message: "failed to parse", Cause: err}
	}
	return New[T](m, opts...)
}

// MergeJSON parses JSON data and merges it with the target struct.
func MergeJSON(target interface{}, data []byte, opts ...*Options) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return &ConversionError{Type: "JSON", Message: "failed to parse", Cause: err}
	}
	return Merge(target, m, opts...)
}

// MergeYAML parses YAML data and merges it with the target struct.
func MergeYAML(target interface{}, data []byte, opts ...*Options) error {
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return &ConversionError{Type: "YAML", Message: "failed to parse", Cause: err}
	}
	return Merge(target, m, opts...)
}

// UnbindJSON converts a struct to JSON bytes.
func UnbindJSON(source interface{}, opts ...*Options) ([]byte, error) {
	m, err := Unbind(source, opts...)
	if err != nil {
		return nil, &ConversionError{Message: "failed to unbind source", Cause: err}
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, &ConversionError{Type: "JSON", Message: "failed to marshal", Cause: err}
	}
	return data, nil
}

// UnbindYAML converts a struct to YAML bytes.
func UnbindYAML(source interface{}, opts ...*Options) ([]byte, error) {
	m, err := Unbind(source, opts...)
	if err != nil {
		return nil, &ConversionError{Message: "failed to unbind source", Cause: err}
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, &ConversionError{Type: "YAML", Message: "failed to marshal", Cause: err}
	}
	return data, nil
}

// --- Reader/Writer Layer ---

// BindJSONReader reads JSON from an io.Reader and binds it to the target struct.
func BindJSONReader(target interface{}, r io.Reader, opts ...*Options) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return &ConversionError{Message: "failed to read from reader", Cause: err}
	}
	return BindJSON(target, data, opts...)
}

// BindYAMLReader reads YAML from an io.Reader and binds it to the target struct.
func BindYAMLReader(target interface{}, r io.Reader, opts ...*Options) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return &ConversionError{Message: "failed to read from reader", Cause: err}
	}
	return BindYAML(target, data, opts...)
}

// NewJSONReader reads JSON from an io.Reader and returns a new instance of type T.
func NewJSONReader[T any](r io.Reader, opts ...*Options) (*T, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, &ConversionError{Message: "failed to read from reader", Cause: err}
	}
	return NewJSON[T](data, opts...)
}

// NewYAMLReader reads YAML from an io.Reader and returns a new instance of type T.
func NewYAMLReader[T any](r io.Reader, opts ...*Options) (*T, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, &ConversionError{Message: "failed to read from reader", Cause: err}
	}
	return NewYAML[T](data, opts...)
}

// MergeJSONReader reads JSON from an io.Reader and merges it with the target struct.
func MergeJSONReader(target interface{}, r io.Reader, opts ...*Options) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return &ConversionError{Message: "failed to read from reader", Cause: err}
	}
	return MergeJSON(target, data, opts...)
}

// MergeYAMLReader reads YAML from an io.Reader and merges it with the target struct.
func MergeYAMLReader(target interface{}, r io.Reader, opts ...*Options) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return &ConversionError{Message: "failed to read from reader", Cause: err}
	}
	return MergeYAML(target, data, opts...)
}

// UnbindJSONWriter converts a struct to JSON and writes it to an io.Writer.
func UnbindJSONWriter(source interface{}, w io.Writer, opts ...*Options) error {
	data, err := UnbindJSON(source, opts...)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return &ConversionError{Message: "failed to write to writer", Cause: err}
	}
	return nil
}

// UnbindYAMLWriter converts a struct to YAML and writes it to an io.Writer.
func UnbindYAMLWriter(source interface{}, w io.Writer, opts ...*Options) error {
	data, err := UnbindYAML(source, opts...)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return &ConversionError{Message: "failed to write to writer", Cause: err}
	}
	return nil
}

// --- File Layer ---

// BindJSONFile reads JSON from the specified file path and binds it to the target struct.
func BindJSONFile(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read JSON", Cause: err}
	}
	return BindJSON(target, data, opts...)
}

// BindYAMLFile reads YAML from the specified file path and binds it to the target struct.
func BindYAMLFile(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read YAML", Cause: err}
	}
	return BindYAML(target, data, opts...)
}

// NewJSONFile reads JSON from the specified file path and returns a new instance of type T.
func NewJSONFile[T any](path string, opts ...*Options) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &FileError{Path: path, Operation: "read JSON", Cause: err}
	}
	return NewJSON[T](data, opts...)
}

// NewYAMLFile reads YAML from the specified file path and returns a new instance of type T.
func NewYAMLFile[T any](path string, opts ...*Options) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &FileError{Path: path, Operation: "read YAML", Cause: err}
	}
	return NewYAML[T](data, opts...)
}

// MergeJSONFile reads JSON from the specified file path and merges it with the target struct.
func MergeJSONFile(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read JSON", Cause: err}
	}
	return MergeJSON(target, data, opts...)
}

// MergeYAMLFile reads YAML from the specified file path and merges it with the target struct.
func MergeYAMLFile(target interface{}, path string, opts ...*Options) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileError{Path: path, Operation: "read YAML", Cause: err}
	}
	return MergeYAML(target, data, opts...)
}

// UnbindJSONFile converts a struct to JSON and writes it to the specified file path.
func UnbindJSONFile(source interface{}, path string, opts ...*Options) error {
	data, err := UnbindJSON(source, opts...)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return &FileError{Path: path, Operation: "write JSON", Cause: err}
	}
	return nil
}

// UnbindYAMLFile converts a struct to YAML and writes it to the specified file path.
func UnbindYAMLFile(source interface{}, path string, opts ...*Options) error {
	data, err := UnbindYAML(source, opts...)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return &FileError{Path: path, Operation: "write YAML", Cause: err}
	}
	return nil
}
