package dd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

// Strict returns Options enabling strict acceptance mode.
//
// dd's default posture is forgiving — unknown keys are ignored, duplicate
// JSON keys resolve last-wins inside the parser, and values coerce across
// types ("5" becomes an int, 5 becomes a string). forgiving YAML retains
// yaml.v3's existing duplicate-key rejection. that is right for config files
// and local records. strict mode is the opposite posture, for data whose
// exact spelling is the contract (signed payloads, hash-pinned documents):
//
//   - intake (BindJSON/BindYAML and their reader/file variants) rejects
//     duplicate keys, trailing data, YAML aliases, and non-JSON scalars, and
//     preserves numbers as json.Number — see DecodeStrictJSON and
//     DecodeStrictYAML, which are also usable directly when a tree needs
//     inspection or normalization between intake and binding
//   - binding rejects input keys the target struct does not declare (unless
//     the struct carries a +extra field, which captures them by declared
//     intent) and refuses type coercion: a number arriving at a string field
//     is an error, not a conversion; map keys must have string as their
//     underlying type
//   - a field tagged +opaque (a map[string]any) accepts any members and
//     captures the raw subtree uninterpreted — syntactic intake rules still
//     apply inside it, binding rules do not
//
// custom Converters, Dynamic binders, and UnmarshalDd implementations remain
// in effect under strict mode. strict intake checks the syntax before
// delegation, but the custom machinery owns what it accepts; a permissive
// Unmarshaler should not be used at an exact contract boundary. Merge does not
// support strict mode.
func Strict() *Options {
	return &Options{Strict: true}
}

// integerLexeme is the strict grammar for a JSON number binding to an
// integer field: no fraction, no exponent.
var integerLexeme = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)

// strictConvertAndSet is the zero-coercion counterpart of convertAndSet: the
// value must be exactly the type the field names. json.Number is the number
// representation strict intake produces and binds to numeric fields with
// exactness checks; native Go values bind to fields of the same kind only.
func strictConvertAndSet(dst reflect.Value, raw interface{}, path string) error {
	// time.Duration is an int64 alias: its defined encoding is the duration
	// string, and nothing else is accepted.
	if dst.Type() == reflect.TypeOf(time.Duration(0)) {
		s, ok := raw.(string)
		if !ok {
			return &TypeMismatchError{Path: path, Expected: "duration string", Actual: fmt.Sprintf("%T", raw)}
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return &ConversionError{Path: path, Value: s, Type: "duration", Cause: err}
		}
		dst.SetInt(int64(d))
		return nil
	}

	// time.Time: the defined encoding is an RFC3339/RFC3339Nano string.
	if dst.Type() == reflect.TypeOf(time.Time{}) {
		v, ok := raw.(string)
		if !ok {
			return &TypeMismatchError{Path: path, Expected: "time (RFC3339 string)", Actual: fmt.Sprintf("%T", raw)}
		}
		t, err := time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return &ConversionError{Path: path, Value: v, Type: "time", Cause: err}
		}
		dst.Set(reflect.ValueOf(t))
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		if n, isNumber := raw.(json.Number); isNumber {
			return &TypeMismatchError{Path: path, Expected: "string", Actual: fmt.Sprintf("number %s", n)}
		}
		rawValue := reflect.ValueOf(raw)
		if rawValue.Kind() != reflect.String {
			return &TypeMismatchError{Path: path, Expected: "string", Actual: fmt.Sprintf("%T", raw)}
		}
		dst.SetString(rawValue.String())
		return nil

	case reflect.Bool:
		rawValue := reflect.ValueOf(raw)
		if !rawValue.IsValid() || rawValue.Kind() != reflect.Bool {
			return &TypeMismatchError{Path: path, Expected: "bool", Actual: fmt.Sprintf("%T", raw)}
		}
		dst.SetBool(rawValue.Bool())
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strictInt64(raw, path)
		if err != nil {
			return err
		}
		if dst.OverflowInt(i64) {
			return &ConversionError{Path: path, Value: fmt.Sprintf("%d", i64), Type: "integer", Message: fmt.Sprintf("value %d overflows %s", i64, dst.Type())}
		}
		dst.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u64, err := strictUint64(raw, path)
		if err != nil {
			return err
		}
		if dst.OverflowUint(u64) {
			return &ConversionError{Path: path, Value: fmt.Sprintf("%d", u64), Type: "unsigned integer", Message: fmt.Sprintf("value %d overflows %s", u64, dst.Type())}
		}
		dst.SetUint(u64)
		return nil

	case reflect.Float32, reflect.Float64:
		f64, err := strictFloat64(raw, path)
		if err != nil {
			return err
		}
		if dst.OverflowFloat(f64) {
			return &ConversionError{Path: path, Type: "float", Message: fmt.Sprintf("value overflows %s", dst.Type())}
		}
		dst.SetFloat(f64)
		return nil

	default:
		return &UnsupportedError{Path: path, Operation: fmt.Sprintf("strict bindings for kind %s", dst.Kind())}
	}
}

// strictInt64 extracts an exact integer: a json.Number with an integer
// lexeme, or a native Go integer. floats and strings are refused.
func strictInt64(raw interface{}, path string) (int64, error) {
	if n, ok := raw.(json.Number); ok {
		if !integerLexeme.MatchString(string(n)) {
			return 0, &ConversionError{Path: path, Value: string(n), Type: "integer", Message: fmt.Sprintf("number %s is not an integer", n)}
		}
		i, err := strconv.ParseInt(string(n), 10, 64)
		if err != nil {
			return 0, &ConversionError{Path: path, Value: string(n), Type: "integer", Cause: err}
		}
		return i, nil
	}
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rawValue.Int(), nil
	default:
		return 0, &TypeMismatchError{Path: path, Expected: "integer", Actual: fmt.Sprintf("%T", raw)}
	}
}

// strictUint64 extracts an exact unsigned integer: a non-negative json.Number
// with an integer lexeme, or a native Go unsigned integer. signed values,
// floats, and strings are refused.
func strictUint64(raw interface{}, path string) (uint64, error) {
	if n, ok := raw.(json.Number); ok {
		lexeme := string(n)
		if !integerLexeme.MatchString(lexeme) {
			return 0, &ConversionError{Path: path, Value: lexeme, Type: "unsigned integer", Message: fmt.Sprintf("number %s is not an integer", n)}
		}
		if lexeme[0] == '-' {
			return 0, &ConversionError{Path: path, Value: lexeme, Type: "unsigned integer", Message: "negative value for unsigned field"}
		}
		u, err := strconv.ParseUint(lexeme, 10, 64)
		if err != nil {
			return 0, &ConversionError{Path: path, Value: lexeme, Type: "unsigned integer", Cause: err}
		}
		return u, nil
	}
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rawValue.Uint(), nil
	default:
		return 0, &TypeMismatchError{Path: path, Expected: "unsigned integer", Actual: fmt.Sprintf("%T", raw)}
	}
}

// strictFloat64 extracts an exact float: any json.Number, or a native Go
// float. integers-as-Go-values, strings, and everything else are refused.
func strictFloat64(raw interface{}, path string) (float64, error) {
	if n, ok := raw.(json.Number); ok {
		f, err := n.Float64()
		if err != nil {
			return 0, &ConversionError{Path: path, Value: string(n), Type: "float", Cause: err}
		}
		return f, nil
	}
	rawValue := reflect.ValueOf(raw)
	switch rawValue.Kind() {
	case reflect.Float32, reflect.Float64:
		return rawValue.Float(), nil
	default:
		return 0, &TypeMismatchError{Path: path, Expected: "float", Actual: fmt.Sprintf("%T", raw)}
	}
}
