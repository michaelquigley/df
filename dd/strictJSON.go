package dd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// DecodeStrictJSON parses JSON bytes into a map[string]any tree under the
// strict acceptance rules: duplicate member names are rejected anywhere
// (JSON parsers legally disagree on which duplicate wins, so one document
// must not have two meanings), trailing data after the document is rejected,
// the top level must be an object, and numbers are preserved as json.Number
// rather than collapsing through float64.
//
// the returned tree is the same shape the forgiving intake produces, so it
// can be inspected or normalized before binding with Bind(target, m,
// dd.Strict()). BindJSON with the Strict option uses this decoder
// automatically.
func DecodeStrictJSON(data []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	value, err := decodeStrictJSONValue(dec, "$")
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, &ConversionError{Type: "JSON", Path: "$", Message: "trailing data after document"}
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil, &TypeMismatchError{Path: "$", Expected: "object at top level", Actual: fmt.Sprintf("%T", value)}
	}
	return m, nil
}

func decodeStrictJSONValue(dec *json.Decoder, path string) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, &ConversionError{Type: "JSON", Path: path, Message: "failed to parse", Cause: err}
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return decodeStrictJSONObject(dec, path)
		case '[':
			return decodeStrictJSONArray(dec, path)
		default:
			return nil, &ConversionError{Type: "JSON", Path: path, Message: fmt.Sprintf("unexpected %q", t)}
		}
	case string, bool, json.Number, nil:
		return t, nil
	default:
		return nil, &ConversionError{Type: "JSON", Path: path, Message: fmt.Sprintf("unexpected token %v", tok)}
	}
}

func decodeStrictJSONObject(dec *json.Decoder, path string) (map[string]any, error) {
	obj := map[string]any{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, &ConversionError{Type: "JSON", Path: path, Message: "failed to parse", Cause: err}
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, &ConversionError{Type: "JSON", Path: path, Message: fmt.Sprintf("non-string member name %v", keyTok)}
		}
		if _, dup := obj[key]; dup {
			return nil, &DuplicateKeyError{Path: path, Key: key}
		}
		value, err := decodeStrictJSONValue(dec, path+"."+key)
		if err != nil {
			return nil, err
		}
		obj[key] = value
	}
	if _, err := dec.Token(); err != nil { // closing '}'
		return nil, &ConversionError{Type: "JSON", Path: path, Message: "failed to parse", Cause: err}
	}
	return obj, nil
}

func decodeStrictJSONArray(dec *json.Decoder, path string) ([]any, error) {
	arr := []any{}
	for dec.More() {
		value, err := decodeStrictJSONValue(dec, fmt.Sprintf("%s[%d]", path, len(arr)))
		if err != nil {
			return nil, err
		}
		arr = append(arr, value)
	}
	if _, err := dec.Token(); err != nil { // closing ']'
		return nil, &ConversionError{Type: "JSON", Path: path, Message: "failed to parse", Cause: err}
	}
	return arr, nil
}
