package df

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func convertAndSet(dst reflect.Value, raw interface{}, path string, opt *Options) error {
	// check for custom converter first
	if converted, wasConverted, err := tryCustomConverter(dst.Type(), raw, opt, true); err != nil {
		return &ConversionError{Path: path, Cause: err}
	} else if wasConverted {
		dst.Set(reflect.ValueOf(converted))
		return nil
	}

	dstKind := dst.Kind()
	// special-case time.Duration (which is an int64 alias)
	if dst.Type() == reflect.TypeOf(time.Duration(0)) {
		switch v := raw.(type) {
		case string:
			d, err := time.ParseDuration(v)
			if err != nil {
				return &ConversionError{Path: path, Value: v, Type: "duration", Cause: err}
			}
			dst.SetInt(int64(d))
			return nil
		case int, int32, int64:
			dst.SetInt(reflect.ValueOf(v).Int())
			return nil
		case float32, float64:
			dst.SetInt(int64(reflect.ValueOf(v).Float()))
			return nil
		default:
			return &TypeMismatchError{Path: path, Expected: "duration (string or number)", Actual: fmt.Sprintf("%T", raw)}
		}
	}

	switch dstKind {
	case reflect.String:
		// handle both string and custom string types
		switch v := raw.(type) {
		case string:
			dst.SetString(v)
			return nil
		default:
			// check if raw value is also a string-based custom type
			rawValue := reflect.ValueOf(raw)
			if rawValue.Kind() == reflect.String {
				dst.SetString(rawValue.String())
				return nil
			}
			return &TypeMismatchError{Path: path, Expected: "string", Actual: fmt.Sprintf("%T", raw)}
		}

	case reflect.Bool:
		switch v := raw.(type) {
		case bool:
			dst.SetBool(v)
			return nil
		case string:
			b, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return &ConversionError{Path: path, Value: v, Type: "bool", Message: fmt.Sprintf("cannot parse bool %q", v)}
			}
			dst.SetBool(b)
			return nil
		default:
			// check if raw value is a bool-based custom type
			rawValue := reflect.ValueOf(raw)
			if rawValue.Kind() == reflect.Bool {
				dst.SetBool(rawValue.Bool())
				return nil
			}
			return &TypeMismatchError{Path: path, Expected: "bool", Actual: fmt.Sprintf("%T", raw)}
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, ok := coerceToInt64(raw)
		if !ok {
			return &TypeMismatchError{Path: path, Expected: "integer", Actual: fmt.Sprintf("%T", raw)}
		}
		dst.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u64, ok := coerceToUint64(raw)
		if !ok {
			return &TypeMismatchError{Path: path, Expected: "unsigned integer", Actual: fmt.Sprintf("%T", raw)}
		}
		dst.SetUint(u64)
		return nil

	case reflect.Float32, reflect.Float64:
		f64, ok := coerceToFloat64(raw)
		if !ok {
			return &TypeMismatchError{Path: path, Expected: "float", Actual: fmt.Sprintf("%T", raw)}
		}
		dst.SetFloat(f64)
		return nil
	}

	return &UnsupportedError{Path: path, Type: fmt.Sprintf("kind %s", dstKind)}
}

func coerceToInt64(raw interface{}) (int64, bool) {
	switch v := raw.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return int64(reflect.ValueOf(v).Uint()), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		if v == "" {
			return 0, false
		}
		if i, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return int64(f), true
		}
	default:
		// handle custom integer types using reflection
		rawValue := reflect.ValueOf(raw)
		switch rawValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rawValue.Int(), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return int64(rawValue.Uint()), true
		}
	}
	return 0, false
}

func coerceToUint64(raw interface{}) (uint64, bool) {
	switch v := raw.(type) {
	case uint:
		return uint64(v), true
	case uint8:
		return uint64(v), true
	case uint16:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case uintptr:
		return uint64(v), true
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(v).Int()
		if i < 0 {
			return 0, false
		}
		return uint64(i), true
	case float32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case float64:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case string:
		if v == "" {
			return 0, false
		}
		if u, err := strconv.ParseUint(strings.TrimSpace(v), 10, 64); err == nil {
			return u, true
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && f >= 0 {
			return uint64(f), true
		}
	default:
		// handle custom unsigned integer types using reflection
		rawValue := reflect.ValueOf(raw)
		switch rawValue.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return rawValue.Uint(), true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i := rawValue.Int()
			if i < 0 {
				return 0, false
			}
			return uint64(i), true
		}
	}
	return 0, false
}

func coerceToFloat64(raw interface{}) (float64, bool) {
	switch v := raw.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), true
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return float64(reflect.ValueOf(v).Uint()), true
	case string:
		if v == "" {
			return 0, false
		}
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	default:
		// handle custom float types using reflection
		rawValue := reflect.ValueOf(raw)
		switch rawValue.Kind() {
		case reflect.Float32, reflect.Float64:
			return rawValue.Float(), true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rawValue.Int()), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return float64(rawValue.Uint()), true
		}
	}
	return 0, false
}