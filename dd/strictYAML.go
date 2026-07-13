package dd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"gopkg.in/yaml.v3"
)

// jsonNumberLexeme matches scalar lexemes that are already valid JSON
// numbers, so an authored "5.00" survives to binding with its trailing zeros
// intact instead of collapsing through a float.
var jsonNumberLexeme = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// DecodeStrictYAML parses YAML bytes into a map[string]any tree under the
// strict acceptance rules: duplicate mapping keys are rejected, aliases and
// anchors are rejected, mapping keys must be plain strings, the document must
// be a single mapping, and numbers are preserved as json.Number carrying the
// authored lexeme where it is JSON-valid (a bare 5.00 stays "5.00").
//
// scalars outside the JSON value model — timestamps, binary — are rejected;
// quote them to bind as strings. the returned tree is the same shape the
// forgiving intake produces. BindYAML with the Strict option uses this
// decoder automatically.
func DecodeStrictYAML(data []byte) (map[string]any, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var root yaml.Node
	if err := dec.Decode(&root); err != nil {
		return nil, &ConversionError{Type: "YAML", Message: "failed to parse", Cause: err}
	}
	// yaml silently reads only the first document; probing for a second is
	// what makes multi-document input a loud rejection.
	var second yaml.Node
	if err := dec.Decode(&second); !errors.Is(err, io.EOF) {
		return nil, &ConversionError{Type: "YAML", Path: "$", Message: "multi-document YAML is rejected"}
	}
	if root.Kind == 0 || len(root.Content) == 0 {
		return nil, &ConversionError{Type: "YAML", Path: "$", Message: "empty document"}
	}
	value, err := strictYAMLValue(root.Content[0], "$")
	if err != nil {
		return nil, err
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil, &TypeMismatchError{Path: "$", Expected: "mapping at top level", Actual: fmt.Sprintf("%T", value)}
	}
	return m, nil
}

func strictYAMLValue(n *yaml.Node, path string) (any, error) {
	if n.Anchor != "" {
		return nil, &ConversionError{Type: "YAML", Path: path, Message: "anchors are rejected in strict mode"}
	}
	switch n.Kind {
	case yaml.MappingNode:
		obj := map[string]any{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			keyNode, valNode := n.Content[i], n.Content[i+1]
			if keyNode.Anchor != "" {
				return nil, &ConversionError{Type: "YAML", Path: path, Message: "anchors are rejected in strict mode"}
			}
			if keyNode.Kind != yaml.ScalarNode || keyNode.Tag != "!!str" {
				return nil, &ConversionError{Type: "YAML", Path: path, Message: "mapping keys must be plain strings"}
			}
			key := keyNode.Value
			if _, dup := obj[key]; dup {
				return nil, &DuplicateKeyError{Path: path, Key: key}
			}
			value, err := strictYAMLValue(valNode, path+"."+key)
			if err != nil {
				return nil, err
			}
			obj[key] = value
		}
		return obj, nil
	case yaml.SequenceNode:
		arr := []any{}
		for i, item := range n.Content {
			value, err := strictYAMLValue(item, fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return nil, err
			}
			arr = append(arr, value)
		}
		return arr, nil
	case yaml.ScalarNode:
		return strictYAMLScalar(n, path)
	case yaml.AliasNode:
		return nil, &ConversionError{Type: "YAML", Path: path, Message: "aliases are rejected in strict mode"}
	default:
		return nil, &ConversionError{Type: "YAML", Path: path, Message: "unsupported node kind"}
	}
}

func strictYAMLScalar(n *yaml.Node, path string) (any, error) {
	switch n.Tag {
	case "!!str":
		return n.Value, nil
	case "!!bool":
		var b bool
		if err := n.Decode(&b); err != nil {
			return nil, &ConversionError{Type: "YAML", Path: path, Message: "failed to parse bool", Cause: err}
		}
		return b, nil
	case "!!null":
		return nil, nil
	case "!!int":
		if !jsonNumberLexeme.MatchString(n.Value) {
			return nil, &ConversionError{Type: "YAML", Path: path, Message: fmt.Sprintf("integer %q is not expressible as a JSON number", n.Value)}
		}
		return json.Number(n.Value), nil
	case "!!float":
		if jsonNumberLexeme.MatchString(n.Value) {
			return json.Number(n.Value), nil
		}
		return nil, &ConversionError{Type: "YAML", Path: path, Message: fmt.Sprintf("float %q is not expressible as a JSON number", n.Value)}
	default:
		return nil, &ConversionError{Type: "YAML", Path: path, Message: fmt.Sprintf("scalar tag %s is rejected in strict mode — quote the value to bind it as a string", n.Tag)}
	}
}
