package dd

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// InspectOptions configures inspection behavior.
type InspectOptions struct {
	// MaxDepth limits recursion depth to prevent infinite loops.
	MaxDepth int
	// Indent sets the indentation string (defaults to "  ").
	Indent string
	// ShowSecrets includes secret fields in output when true.
	ShowSecrets bool
}

// Inspect returns a human-readable representation of a struct's resolved state.
// designed for configuration debugging and validation. secret fields marked with
// `dd:",+secret"` are hidden unless ShowSecrets is true.
//
// the output format is a clean, indented pseudo-data structure optimized for
// readability rather than parseability.
//
// supported types:
// - primitives: string, bool, all int/uint sizes, float32/64, time.Duration
// - pointers to the above (nil pointers shown as "<nil>")
// - structs and pointers to structs (recursively inspected)
// - slices of the above (shown as numbered lists)
// - Dynamic interface implementations (shown with their type)
// - Pointer[T] references (shown with resolved state)
//
// opts are optional; pass nil or omit to use defaults.
func Inspect(source interface{}, opts ...*InspectOptions) (string, error) {
	if source == nil {
		return "<nil>", nil
	}

	opt := getInspectOptions(opts...)

	val := reflect.ValueOf(source)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "<nil>", nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", &TypeMismatchError{Expected: "struct or pointer to struct", Actual: fmt.Sprintf("%T", source)}
	}

	// first pass: calculate the maximum field name length and depth across all structures
	maxNameLength := calculateMaxFieldNameLength(val, 0, opt)
	maxDepth := calculateMaxDepth(val, 0, opt)

	// calculate the global colon position: max depth * indent + max field name + space before colon
	globalColonPos := maxDepth*len(opt.Indent) + maxNameLength

	var builder strings.Builder
	if err := inspectStructWithAlignment(val, &builder, 0, opt, globalColonPos); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// MustInspect returns a human-readable representation of a struct's resolved state,
// panicking if an error occurs. see Inspect for full documentation.
func MustInspect(source interface{}, opts ...*InspectOptions) string {
	out, err := Inspect(source, opts...)
	if err != nil {
		panic(err)
	}
	return out
}

func getInspectOptions(opts ...*InspectOptions) *InspectOptions {
	if len(opts) == 0 || opts[0] == nil {
		return &InspectOptions{
			MaxDepth:    10,
			Indent:      "  ",
			ShowSecrets: false,
		}
	}
	opt := *opts[0]
	if opt.MaxDepth <= 0 {
		opt.MaxDepth = 10
	}
	if opt.Indent == "" {
		opt.Indent = "  "
	}
	return &opt
}

func calculateMaxDepth(val reflect.Value, depth int, opt *InspectOptions) int {
	if depth > opt.MaxDepth {
		return depth
	}

	maxDepth := depth

	// handle different value types
	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			return calculateMaxDepth(val.Elem(), depth, opt)
		}
		return depth
	case reflect.Struct:
		maxDepth = max(maxDepth, calculateMaxDepthFromStruct(val, depth, opt))
	case reflect.Slice:
		if !val.IsNil() {
			elemType := val.Type().Elem()
			if elemType.Kind() == reflect.Struct || (elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct) {
				for i := 0; i < val.Len(); i++ {
					maxDepth = max(maxDepth, calculateMaxDepth(val.Index(i), depth+1, opt))
				}
			}
		}
	}

	return maxDepth
}

func calculateMaxDepthFromStruct(structVal reflect.Value, depth int, opt *InspectOptions) int {
	if depth > opt.MaxDepth {
		return depth
	}

	structType := structVal.Type()
	maxDepth := depth

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		fieldVal := structVal.Field(i)

		// handle embedded structs by recursively calculating their depth
		if field.Anonymous {
			var embeddedVal reflect.Value
			if field.Type.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					continue // skip nil embedded pointer
				}
				embeddedVal = fieldVal.Elem()
			} else {
				embeddedVal = fieldVal
			}

			if embeddedVal.Kind() == reflect.Struct {
				maxDepth = max(maxDepth, calculateMaxDepth(embeddedVal, depth, opt))
			}
			continue
		}

		tag := parseDdTag(field)
		if tag.Skip {
			continue
		}

		// recursively check nested structures
		maxDepth = max(maxDepth, calculateMaxDepth(fieldVal, depth+1, opt))
	}

	return maxDepth
}

func calculateMaxFieldNameLength(val reflect.Value, depth int, opt *InspectOptions) int {
	if depth > opt.MaxDepth {
		return 0
	}

	maxLength := 0

	// handle different value types
	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			return calculateMaxFieldNameLength(val.Elem(), depth, opt)
		}
		return 0
	case reflect.Struct:
		maxLength = max(maxLength, calculateMaxFieldNameLengthFromStruct(val, depth, opt))
	case reflect.Slice:
		if !val.IsNil() {
			elemType := val.Type().Elem()
			if elemType.Kind() == reflect.Struct || (elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct) {
				for i := 0; i < val.Len(); i++ {
					maxLength = max(maxLength, calculateMaxFieldNameLength(val.Index(i), depth+1, opt))
				}
			}
		}
	}

	return maxLength
}

func calculateMaxFieldNameLengthFromStruct(structVal reflect.Value, depth int, opt *InspectOptions) int {
	if depth > opt.MaxDepth {
		return 0
	}

	structType := structVal.Type()
	maxLength := 0

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		fieldVal := structVal.Field(i)

		// handle embedded structs by recursively calculating their field name lengths
		if field.Anonymous {
			var embeddedVal reflect.Value
			if field.Type.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					continue // skip nil embedded pointer
				}
				embeddedVal = fieldVal.Elem()
			} else {
				embeddedVal = fieldVal
			}

			if embeddedVal.Kind() == reflect.Struct {
				maxLength = max(maxLength, calculateMaxFieldNameLength(embeddedVal, depth, opt))
			}
			continue
		}

		tag := parseDdTag(field)
		if tag.Skip {
			continue
		}

		name := tag.Name
		if name == "" {
			name = toSnakeCase(field.Name)
		}

		// calculate display name with secret annotation
		displayName := name
		if tag.Secret {
			displayName += " (secret)"
		}

		maxLength = max(maxLength, len(displayName))

		// recursively check nested structures
		maxLength = max(maxLength, calculateMaxFieldNameLength(fieldVal, depth+1, opt))
	}

	return maxLength
}

func inspectStructWithAlignment(structVal reflect.Value, builder *strings.Builder, depth int, opt *InspectOptions, globalColonPos int) error {
	if depth > opt.MaxDepth {
		builder.WriteString("<max depth reached>")
		return nil
	}

	structType := structVal.Type()
	typeName := structType.Name()
	if typeName == "" {
		typeName = "struct"
	}

	builder.WriteString(typeName)
	builder.WriteString(" {\n")

	// collect field info
	type fieldInfo struct {
		name        string
		tag         DdTag
		fieldVal    reflect.Value
		displayName string
	}

	var fields []fieldInfo

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}

		fieldVal := structVal.Field(i)

		// handle embedded structs by flattening their fields into the parent
		if field.Anonymous {
			var embeddedVal reflect.Value
			if field.Type.Kind() == reflect.Ptr {
				if fieldVal.IsNil() {
					continue // skip nil embedded pointer
				}
				embeddedVal = fieldVal.Elem()
			} else {
				embeddedVal = fieldVal
			}

			if embeddedVal.Kind() == reflect.Struct {
				// recursively collect embedded struct fields
				embeddedType := embeddedVal.Type()
				for j := 0; j < embeddedVal.NumField(); j++ {
					embeddedField := embeddedType.Field(j)
					if embeddedField.PkgPath != "" { // unexported
						continue
					}

					embeddedTag := parseDdTag(embeddedField)
					if embeddedTag.Skip {
						continue
					}

					embeddedName := embeddedTag.Name
					if embeddedName == "" {
						embeddedName = toSnakeCase(embeddedField.Name)
					}

					embeddedFieldVal := embeddedVal.Field(j)

					// calculate display name with secret annotation
					embeddedDisplayName := embeddedName
					if embeddedTag.Secret {
						embeddedDisplayName += " (secret)"
					}

					fields = append(fields, fieldInfo{
						name:        embeddedName,
						tag:         embeddedTag,
						fieldVal:    embeddedFieldVal,
						displayName: embeddedDisplayName,
					})
				}
			}
			continue
		}

		tag := parseDdTag(field)
		if tag.Skip {
			continue
		}
		name := tag.Name
		if name == "" {
			name = toSnakeCase(field.Name)
		}

		// calculate display name with secret annotation
		displayName := name
		if tag.Secret {
			displayName += " (secret)"
		}

		fields = append(fields, fieldInfo{
			name:        name,
			tag:         tag,
			fieldVal:    fieldVal,
			displayName: displayName,
		})
	}

	hasFields := len(fields) > 0
	for _, f := range fields {
		// write indentation
		for j := 0; j <= depth; j++ {
			builder.WriteString(opt.Indent)
		}

		// write field name with padding for GLOBAL alignment
		builder.WriteString(f.displayName)

		// calculate current position: indentation + field name length
		currentPos := (depth+1)*len(opt.Indent) + len(f.displayName)

		// pad to reach the global colon position
		padding := globalColonPos - currentPos
		for k := 0; k < padding; k++ {
			builder.WriteString(" ")
		}
		builder.WriteString(": ")

		if f.tag.Secret && !opt.ShowSecrets {
			// show <set> or <unset> instead of actual value
			if isSecretFieldEmpty(f.fieldVal) {
				builder.WriteString("<unset>")
			} else {
				builder.WriteString("<set>")
			}
		} else {
			if err := inspectValueWithAlignment(f.fieldVal, builder, depth+1, opt, globalColonPos); err != nil {
				return err
			}
		}

		builder.WriteString("\n")
	}

	if !hasFields {
		for j := 0; j <= depth; j++ {
			builder.WriteString(opt.Indent)
		}
		builder.WriteString("<no fields>")
		builder.WriteString("\n")
	}

	// write closing brace indentation
	for j := 0; j < depth; j++ {
		builder.WriteString(opt.Indent)
	}
	builder.WriteString("}")

	return nil
}

func inspectValueWithAlignment(val reflect.Value, builder *strings.Builder, depth int, opt *InspectOptions, globalColonPos int) error {
	if depth > opt.MaxDepth {
		builder.WriteString("<max depth reached>")
		return nil
	}

	// handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			builder.WriteString("<nil>")
			return nil
		}
		return inspectValueWithAlignment(val.Elem(), builder, depth, opt, globalColonPos)
	}

	// check for Pointer[T] type
	if isPointerType(val.Type()) {
		return inspectPointerTypeWithAlignment(val, builder, depth, opt, globalColonPos)
	}

	// check for Dynamic interface
	if val.Type() == dynamicInterfaceType {
		if val.IsNil() {
			builder.WriteString("<nil Dynamic>")
			return nil
		}
		dynVal := val.Interface().(Dynamic)
		builder.WriteString(dynVal.Type())
		return nil
	}

	switch val.Kind() {
	case reflect.String:
		builder.WriteString(strconv.Quote(val.String()))
	case reflect.Bool:
		builder.WriteString(strconv.FormatBool(val.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// handle time.Duration specially
		if val.Type() == reflect.TypeOf(time.Duration(0)) {
			dur := time.Duration(val.Int())
			builder.WriteString(dur.String())
		} else {
			builder.WriteString(strconv.FormatInt(val.Int(), 10))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		builder.WriteString(strconv.FormatUint(val.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		builder.WriteString(strconv.FormatFloat(val.Float(), 'g', -1, val.Type().Bits()))
	case reflect.Slice:
		return inspectSliceWithAlignment(val, builder, depth, opt, globalColonPos)
	case reflect.Struct:
		return inspectStructWithAlignment(val, builder, depth, opt, globalColonPos)
	case reflect.Map:
		return inspectMapWithAlignment(val, builder, depth, opt, globalColonPos)
	case reflect.Interface:
		if val.IsNil() {
			builder.WriteString("<nil>")
		} else {
			return inspectValueWithAlignment(val.Elem(), builder, depth, opt, globalColonPos)
		}
	default:
		builder.WriteString(fmt.Sprintf("<%s>", val.Type().String()))
	}

	return nil
}

func inspectSliceWithAlignment(val reflect.Value, builder *strings.Builder, depth int, opt *InspectOptions, globalColonPos int) error {
	if val.IsNil() {
		builder.WriteString("<nil slice>")
		return nil
	}

	if val.Len() == 0 {
		builder.WriteString("[]")
		return nil
	}

	builder.WriteString("[\n")

	for i := 0; i < val.Len(); i++ {
		// write indentation
		for j := 0; j <= depth; j++ {
			builder.WriteString(opt.Indent)
		}

		builder.WriteString(fmt.Sprintf("[%d]: ", i))

		if err := inspectValueWithAlignment(val.Index(i), builder, depth+1, opt, globalColonPos); err != nil {
			return err
		}

		builder.WriteString("\n")
	}

	// write closing bracket indentation
	for j := 0; j < depth; j++ {
		builder.WriteString(opt.Indent)
	}
	builder.WriteString("]")

	return nil
}

func isSecretFieldEmpty(val reflect.Value) bool {
	// handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true
		}
		val = val.Elem()
	}

	// check if the field is empty based on its type
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Slice:
		return val.IsNil() || val.Len() == 0
	case reflect.Struct:
		// for structs, consider empty if it's the zero value
		return val.IsZero()
	case reflect.Interface:
		return val.IsNil()
	default:
		// for other types, use the zero value check
		return val.IsZero()
	}
}

func inspectPointerTypeWithAlignment(val reflect.Value, builder *strings.Builder, depth int, opt *InspectOptions, globalColonPos int) error {
	// get the Ref field
	refField := val.FieldByName("Ref")
	if !refField.IsValid() {
		builder.WriteString("<invalid Pointer>")
		return nil
	}

	ref := refField.String()
	if ref == "" {
		builder.WriteString("<empty ref>")
		return nil
	}

	// get the Resolved field
	resolvedField := val.FieldByName("Resolved")
	if !resolvedField.IsValid() {
		builder.WriteString(fmt.Sprintf("$ref: %s", strconv.Quote(ref)))
		return nil
	}

	builder.WriteString(fmt.Sprintf("$ref: %s -> ", strconv.Quote(ref)))

	// check if resolved
	if resolvedField.Kind() == reflect.Ptr && resolvedField.IsNil() {
		builder.WriteString("<unresolved>")
		return nil
	}

	return inspectValueWithAlignment(resolvedField, builder, depth, opt, globalColonPos)
}

func inspectMapWithAlignment(val reflect.Value, builder *strings.Builder, depth int, opt *InspectOptions, globalColonPos int) error {
	if val.IsNil() {
		builder.WriteString("<nil map>")
		return nil
	}

	if val.Len() == 0 {
		builder.WriteString("{}")
		return nil
	}

	builder.WriteString("{\n")

	keys := val.MapKeys()
	for i, key := range keys {
		// write indentation
		for j := 0; j <= depth; j++ {
			builder.WriteString(opt.Indent)
		}

		// write key
		if key.Kind() == reflect.String {
			builder.WriteString(strconv.Quote(key.String()))
		} else {
			builder.WriteString(fmt.Sprintf("%v", key.Interface()))
		}
		builder.WriteString(": ")

		mapVal := val.MapIndex(key)
		if err := inspectValueWithAlignment(mapVal, builder, depth+1, opt, globalColonPos); err != nil {
			return err
		}

		if i < len(keys)-1 {
			builder.WriteString(",")
		}
		builder.WriteString("\n")
	}

	// write closing brace indentation
	for j := 0; j < depth; j++ {
		builder.WriteString(opt.Indent)
	}
	builder.WriteString("}")

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
