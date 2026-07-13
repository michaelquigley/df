package dd

import (
	"encoding/json"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// --- Bytes Layer (base) ---

// intakeJSON parses JSON bytes into the binding tree, selecting the strict
// decoder when strict mode is requested; the forgiving path is unchanged.
func intakeJSON(data []byte, opts []*Options) (map[string]any, error) {
	if opt, err := getOptions(opts...); err != nil {
		return nil, err
	} else if opt != nil && opt.Strict {
		return DecodeStrictJSON(data)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, &ConversionError{Type: "JSON", Message: "failed to parse", Cause: err}
	}
	return m, nil
}

// intakeYAML parses YAML bytes into the binding tree, selecting the strict
// decoder when strict mode is requested; the forgiving path is unchanged.
func intakeYAML(data []byte, opts []*Options) (map[string]any, error) {
	if opt, err := getOptions(opts...); err != nil {
		return nil, err
	} else if opt != nil && opt.Strict {
		return DecodeStrictYAML(data)
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, &ConversionError{Type: "YAML", Message: "failed to parse", Cause: err}
	}
	return m, nil
}

// BindJSON parses JSON data and binds it to the target struct.
func BindJSON(target interface{}, data []byte, opts ...*Options) error {
	m, err := intakeJSON(data, opts)
	if err != nil {
		return err
	}
	return Bind(target, m, opts...)
}

// BindYAML parses YAML data and binds it to the target struct.
func BindYAML(target interface{}, data []byte, opts ...*Options) error {
	m, err := intakeYAML(data, opts)
	if err != nil {
		return err
	}
	return Bind(target, m, opts...)
}

// NewJSON parses JSON data and returns a new instance of type T.
func NewJSON[T any](data []byte, opts ...*Options) (*T, error) {
	m, err := intakeJSON(data, opts)
	if err != nil {
		return nil, err
	}
	return New[T](m, opts...)
}

// NewYAML parses YAML data and returns a new instance of type T.
func NewYAML[T any](data []byte, opts ...*Options) (*T, error) {
	m, err := intakeYAML(data, opts)
	if err != nil {
		return nil, err
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
//
// output is deterministic: for a given input value the produced bytes are identical
// across runs, processes, and versions. all keys — struct field names and map keys
// alike — are emitted in sorted order (encoding/json sorts object keys
// lexicographically), and slice and array element order is preserved as-is. fields
// captured via `+extra` are interleaved in sorted order with the rest rather than
// appended at the end.
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
//
// output is deterministic: for a given input value the produced bytes are identical
// across runs, processes, and versions. all keys — struct field names and map keys
// alike — are emitted in sorted order (yaml.v3 sorts mapping keys), and slice and
// array element order is preserved as-is. fields captured via `+extra` are
// interleaved in sorted order with the rest rather than appended at the end.
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
// output is deterministic with sorted keys; see UnbindJSON.
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
// output is deterministic with sorted keys; see UnbindYAML.
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
// output is deterministic with sorted keys; see UnbindJSON.
func UnbindJSONFile(source interface{}, path string, opts ...*Options) error {
	opt, err := getOptions(opts...)
	if err != nil {
		return err
	}
	data, err := UnbindJSON(source, opt)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, outputFileMode(path, opt)); err != nil {
		return &FileError{Path: path, Operation: "write JSON", Cause: err}
	}
	return nil
}

// UnbindYAMLFile converts a struct to YAML and writes it to the specified file path.
// output is deterministic with sorted keys; see UnbindYAML.
func UnbindYAMLFile(source interface{}, path string, opts ...*Options) error {
	opt, err := getOptions(opts...)
	if err != nil {
		return err
	}
	data, err := UnbindYAML(source, opt)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, outputFileMode(path, opt)); err != nil {
		return &FileError{Path: path, Operation: "write YAML", Cause: err}
	}
	return nil
}

func outputFileMode(path string, opt *Options) fs.FileMode {
	if opt != nil && opt.File != nil && opt.File.Mode != nil {
		return *opt.File.Mode
	}
	if info, err := os.Stat(path); err == nil {
		return info.Mode().Perm()
	}
	return 0o644
}
