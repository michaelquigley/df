package dd

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ValidationError represents errors in input validation
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// TypeMismatchError represents type conversion errors
type TypeMismatchError struct {
	Path     string
	Expected string
	Actual   string
}

func (e *TypeMismatchError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: expected %s, got %s", e.Path, e.Expected, e.Actual)
	}
	return fmt.Sprintf("expected %s, got %s", e.Expected, e.Actual)
}

// ConversionError represents data conversion failures
type ConversionError struct {
	Path    string
	Value   string
	Type    string
	Message string
	Cause   error
}

func (e *ConversionError) Error() string {
	if e.Cause != nil {
		if e.Path != "" {
			return fmt.Sprintf("%s: %s", e.Path, e.Cause.Error())
		}
		return e.Cause.Error()
	}
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Message)
	}
	return e.Message
}

func (e *ConversionError) Unwrap() error {
	return e.Cause
}

// BindingError represents struct field binding errors
type BindingError struct {
	Path  string
	Field string
	Key   string
	Cause error
}

func (e *BindingError) Error() string {
	if e.Cause == nil {
		if e.Key != "" {
			return fmt.Sprintf("binding field %s.%s from key %q", e.Path, e.Field, e.Key)
		}
		return fmt.Sprintf("binding field %s.%s", e.Path, e.Field)
	}
	if e.Key != "" {
		return fmt.Sprintf("binding field %s.%s from key %q: %s", e.Path, e.Field, e.Key, e.Cause.Error())
	}
	return fmt.Sprintf("binding field %s.%s: %s", e.Path, e.Field, e.Cause.Error())
}

func (e *BindingError) Unwrap() error {
	return e.Cause
}

// UnbindingError represents struct field unbinding errors
type UnbindingError struct {
	Path  string
	Field string
	Key   string
	Cause error
}

func (e *UnbindingError) Error() string {
	return fmt.Sprintf("unbinding field %s.%s to key %q: %s", e.Path, e.Field, e.Key, e.Cause.Error())
}

func (e *UnbindingError) Unwrap() error {
	return e.Cause
}

// FileError represents file I/O operation errors
type FileError struct {
	Path      string
	Operation string
	Cause     error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("failed to %s file %s: %s", e.Operation, e.Path, e.Cause.Error())
}

func (e *FileError) Unwrap() error {
	return e.Cause
}

// IsNotFound checks if the FileError represents a file not found error.
func (e *FileError) IsNotFound() bool {
	return errors.Is(e.Cause, os.ErrNotExist)
}

// PointerError represents pointer resolution errors
type PointerError struct {
	Path      string
	Reference string
	Message   string
	Cause     error
}

func (e *PointerError) Error() string {
	if e.Cause != nil {
		if e.Path != "" {
			return fmt.Sprintf("%s: %s", e.Path, e.Cause.Error())
		}
		return e.Cause.Error()
	}
	if e.Reference != "" {
		return fmt.Sprintf("unresolved reference: %s", e.Reference)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Message)
	}
	return e.Message
}

func (e *PointerError) Unwrap() error {
	return e.Cause
}

// UnsupportedError represents unsupported operation errors
type UnsupportedError struct {
	Path      string
	Operation string
	Type      string
}

func (e *UnsupportedError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s are not supported", e.Path, e.Operation)
	}
	return fmt.Sprintf("%s are not supported", e.Operation)
}

// RequiredFieldError represents missing required field errors
type RequiredFieldError struct {
	Path  string
	Field string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("%s.%s: required field missing", e.Path, e.Field)
}

// MultipleExtraFieldsError represents the error when a struct has more than one +extra field
type MultipleExtraFieldsError struct {
	Path string
}

func (e *MultipleExtraFieldsError) Error() string {
	return fmt.Sprintf("%s: multiple +extra fields not allowed", e.Path)
}

// ValueMismatchError represents errors when a field value doesn't match the expected constraint
type ValueMismatchError struct {
	Path     string
	Field    string
	Expected string
	Actual   string
}

func (e *ValueMismatchError) Error() string {
	return fmt.Sprintf("%s.%s: expected value %q, got %q", e.Path, e.Field, e.Expected, e.Actual)
}

// DuplicateKeyError represents a duplicate member name rejected by strict
// intake — parsers legally disagree on which duplicate wins, so one document
// must not have two meanings.
type DuplicateKeyError struct {
	Path string
	Key  string
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf("%s: duplicate key %q", e.Path, e.Key)
}

// UnknownFieldError represents an input key the target struct does not
// declare, rejected under strict binding.
type UnknownFieldError struct {
	Path string
	Key  string
}

func (e *UnknownFieldError) Error() string {
	return fmt.Sprintf("%s: unknown field %q", e.Path, e.Key)
}

// KeyCollisionError represents two distinct map keys that serialize to the
// same string spelling during unbinding. such a map has no lossless JSON or
// YAML form, and which entry survived would depend on go's randomized map
// iteration, so unbind refuses it rather than letting one entry silently win.
type KeyCollisionError struct {
	Key  string   // the spelling the keys collide on
	Keys []string // the colliding source keys with their go types, sorted
}

func (e *KeyCollisionError) Error() string {
	return fmt.Sprintf("map key collision: %s serialize to the same key %q", strings.Join(e.Keys, ", "), e.Key)
}

// IndexError represents errors with array/slice indexing
type IndexError struct {
	Index int
	Cause error
}

func (e *IndexError) Error() string {
	return fmt.Sprintf("index %d: %s", e.Index, e.Cause.Error())
}

func (e *IndexError) Unwrap() error {
	return e.Cause
}
