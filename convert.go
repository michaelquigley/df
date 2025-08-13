package df

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func convertAndSet(dst reflect.Value, raw interface{}, path string) error {
	dstKind := dst.Kind()
	// special-case time.Duration (which is an int64 alias)
	if dst.Type() == reflect.TypeOf(time.Duration(0)) {
		switch v := raw.(type) {
		case string:
			d, err := time.ParseDuration(v)
			if err != nil {
				return fmt.Errorf("%s: invalid duration %q: %w", path, v, err)
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
			return fmt.Errorf("%s: expected duration (string or number), got %T", path, raw)
		}
	}

	switch dstKind {
	case reflect.String:
		s, ok := raw.(string)
		if !ok {
			return fmt.Errorf("%s: expected string, got %T", path, raw)
		}
		dst.SetString(s)
		return nil

	case reflect.Bool:
		switch v := raw.(type) {
		case bool:
			dst.SetBool(v)
			return nil
		case string:
			b, err := strconv.ParseBool(strings.TrimSpace(v))
			if err != nil {
				return fmt.Errorf("%s: cannot parse bool %q", path, v)
			}
			dst.SetBool(b)
			return nil
		default:
			return fmt.Errorf("%s: expected bool, got %T", path, raw)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, ok := coerceToInt64(raw)
		if !ok {
			return fmt.Errorf("%s: expected integer, got %T", path, raw)
		}
		dst.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u64, ok := coerceToUint64(raw)
		if !ok {
			return fmt.Errorf("%s: expected unsigned integer, got %T", path, raw)
		}
		dst.SetUint(u64)
		return nil

	case reflect.Float32, reflect.Float64:
		f64, ok := coerceToFloat64(raw)
		if !ok {
			return fmt.Errorf("%s: expected float, got %T", path, raw)
		}
		dst.SetFloat(f64)
		return nil
	}

	return fmt.Errorf("%s: unsupported kind %s", path, dstKind)
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
	}
	return 0, false
}