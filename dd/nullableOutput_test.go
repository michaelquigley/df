package dd

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNullableOutput(t *testing.T) {
	type nested struct{ Name string }
	text := "stop"
	for _, test := range []struct {
		name    string
		source  any
		present bool
		want    any
	}{
		{"nil primitive pointer", struct {
			Value *string `dd:",+nullable"`
		}{}, true, nil},
		{"nil struct pointer", struct {
			Value *nested `dd:",+nullable"`
		}{}, true, nil},
		{"nil interface", struct {
			Value any `dd:",+nullable"`
		}{}, true, nil},
		{"nil map", struct {
			Value map[string]string `dd:",+nullable"`
		}{}, true, nil},
		{"nil slice", struct {
			Value []string `dd:",+nullable"`
		}{}, true, nil},
		{"typed nil pointer", struct {
			Value any `dd:",+nullable"`
		}{(*nested)(nil)}, true, nil},
		{"typed nil map", struct {
			Value any `dd:",+nullable"`
		}{map[string]string(nil)}, true, nil},
		{"typed nil slice", struct {
			Value any `dd:",+nullable"`
		}{[]string(nil)}, true, nil},
		{"non-nil pointer", struct {
			Value *string `dd:",+nullable"`
		}{&text}, true, "stop"},
		{"non-nil struct", struct {
			Value *nested `dd:",+nullable"`
		}{&nested{Name: "nested"}}, true, map[string]any{"name": "nested"}},
		{"non-nil interface", struct {
			Value any `dd:",+nullable"`
		}{"stop"}, true, "stop"},
		{"non-nil map", struct {
			Value map[string]string `dd:",+nullable"`
		}{map[string]string{"key": "value"}}, true, map[string]any{"key": "value"}},
		{"non-nil slice", struct {
			Value []string `dd:",+nullable"`
		}{[]string{"stop"}}, true, []any{"stop"}},
		{"empty map", struct {
			Value map[string]string `dd:",+nullable"`
		}{map[string]string{}}, true, map[string]any{}},
		{"empty slice", struct {
			Value []string `dd:",+nullable"`
		}{[]string{}}, true, []any{}},
		{"zero string", struct {
			Value string `dd:",+nullable"`
		}{}, true, ""},
		{"zero int", struct {
			Value int `dd:",+nullable"`
		}{}, true, int(0)},
		{"zero bool", struct {
			Value bool `dd:",+nullable"`
		}{}, true, false},
		{"zero struct", struct {
			Value nested `dd:",+nullable"`
		}{}, true, map[string]any{"name": ""}},
		{"untagged nil pointer", struct{ Value *string }{}, false, nil},
		{"untagged nil struct pointer", struct{ Value *nested }{}, false, nil},
		{"untagged nil interface", struct{ Value any }{}, false, nil},
		{"untagged nil map", struct{ Value map[string]string }{}, true, map[string]any{}},
		{"untagged nil slice", struct{ Value []string }{}, true, []any{}},
		{"untagged typed nil pointer", struct{ Value any }{(*string)(nil)}, false, nil},
		{"untagged typed nil map", struct{ Value any }{map[string]string(nil)}, true, map[string]any{}},
		{"untagged typed nil slice", struct{ Value any }{[]string(nil)}, true, []any{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			out, err := Unbind(test.source)
			require.NoError(t, err)
			value, present := out["value"]
			assert.Equal(t, test.present, present)
			assert.Equal(t, test.want, value)
		})
	}
}

func TestNullableOutputOmitEmptyPrecedence(t *testing.T) {
	type document struct {
		Pointer *string           `dd:",+nullable,+omitempty"`
		Object  *struct{}         `dd:",+nullable,+omitempty"`
		Value   any               `dd:",+nullable,+omitempty"`
		Map     map[string]string `dd:",+nullable,+omitempty"`
		Slice   []string          `dd:",+nullable,+omitempty"`
		Number  int               `dd:",+nullable,+omitempty"`
	}
	out, err := Unbind(document{})
	require.NoError(t, err)
	assert.Empty(t, out)

	zero := ""
	out, err = Unbind(document{
		Pointer: &zero,
		Object:  &struct{}{},
		Map:     map[string]string{},
		Slice:   []string{},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"pointer": ""}, out)

	for _, raw := range []any{(*string)(nil), map[string]string(nil), []string(nil)} {
		out, err := Unbind(document{Value: raw})
		require.NoError(t, err)
		value, present := out["value"]
		assert.True(t, present)
		assert.Nil(t, value)
	}
	// a non-nil interface is tested at the field boundary; an empty output map
	// is still omitted by the existing post-conversion check.
	out, err = Unbind(document{Value: map[string]string{}})
	require.NoError(t, err)
	assert.NotContains(t, out, "value")
	out, err = Unbind(document{Value: []string{}})
	require.NoError(t, err)
	assert.Contains(t, out, "value")
	assert.Equal(t, []any{}, out["value"])
}

func TestNullableOutputUnsupportedValues(t *testing.T) {
	for _, raw := range []any{(chan int)(nil), (func())(nil), unsafe.Pointer(nil), make(chan int), func() {}} {
		var unsupported *UnsupportedError
		_, err := Unbind(struct {
			Value any `dd:",+nullable"`
		}{raw})
		assert.ErrorAs(t, err, &unsupported)
	}
	var unsupported *UnsupportedError
	_, err := Unbind(struct {
		Value chan int `dd:",+nullable"`
	}{})
	assert.ErrorAs(t, err, &unsupported)
}

type NullableOutputEmbedded struct {
	Content any `dd:"message,+opaque,+nullable"`
}

func TestNullableOutputTraversal(t *testing.T) {
	out, err := Unbind(struct {
		Nested NullableOutputEmbedded
		NullableOutputEmbedded
		Skipped *string        `dd:"-"`
		hidden  *string        `dd:",+nullable"`
		Extra   map[string]any `dd:",+extra,+nullable"`
	}{})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"nested": map[string]any{"message": nil}, "message": nil}, out)

	type pointerEmbed struct {
		*NullableOutputEmbedded
		Extra map[string]any `dd:",+extra,+nullable"`
	}
	out, err = Unbind(pointerEmbed{})
	require.NoError(t, err)
	assert.Empty(t, out)
	out, err = Unbind(pointerEmbed{
		NullableOutputEmbedded: &NullableOutputEmbedded{},
		Extra:                  map[string]any{"extension": true},
	})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"message": nil, "extension": true}, out)
	_, err = Unbind(pointerEmbed{
		NullableOutputEmbedded: &NullableOutputEmbedded{},
		Extra:                  map[string]any{"message": "collision"},
	})
	var validation *ValidationError
	assert.ErrorAs(t, err, &validation)
}

type nullableOutputMarshaler struct{ Called bool }

func (m *nullableOutputMarshaler) MarshalDd() (map[string]any, error) {
	if m == nil {
		return nil, errors.New("nil nullable marshaler must not be called")
	}
	m.Called = true
	return map[string]any{"custom": true}, nil
}

func TestNullableOutputMarshaler(t *testing.T) {
	type document struct {
		Value *nullableOutputMarshaler `dd:",+nullable"`
	}
	out, err := Unbind(document{})
	require.NoError(t, err)
	assert.Contains(t, out, "value")
	assert.Nil(t, out["value"])
	out, err = Unbind(struct {
		Value any `dd:",+nullable"`
	}{(*nullableOutputMarshaler)(nil)})
	require.NoError(t, err)
	assert.Contains(t, out, "value")
	assert.Nil(t, out["value"])

	marshaler := &nullableOutputMarshaler{}
	out, err = Unbind(document{Value: marshaler})
	require.NoError(t, err)
	assert.True(t, marshaler.Called)
	assert.Equal(t, map[string]any{"custom": true}, out["value"])
	out, err = Unbind(struct {
		Value any `dd:",+opaque,+nullable"`
	}{marshaler})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"custom": true}, out["value"])
}

func TestNullableOutputConverter(t *testing.T) {
	converter := &opaqueConverter{}
	opt := &Options{Converters: map[reflect.Type]Converter{
		reflect.TypeOf((*any)(nil)).Elem(): converter,
	}}
	type document struct {
		Value any `dd:",+nullable"`
	}
	for _, raw := range []any{nil, (*string)(nil), map[string]string(nil), []string(nil)} {
		out, err := Unbind(document{Value: raw}, opt)
		require.NoError(t, err)
		assert.Contains(t, out, "value")
		assert.Nil(t, out["value"])
		assert.False(t, converter.ToCalled)
	}
	out, err := Unbind(document{Value: "stop"}, opt)
	require.NoError(t, err)
	assert.True(t, converter.ToCalled)
	assert.Equal(t, "converted", out["value"])
}

func TestNullableOutputFormats(t *testing.T) {
	type chunk struct {
		FinishReason *string `dd:",+nullable"`
	}
	stop := "stop"
	for _, reason := range []*string{nil, &stop} {
		target := chunk{FinishReason: reason}
		want := any(nil)
		if reason != nil {
			want = *reason
		}
		out, err := Unbind(target)
		require.NoError(t, err)
		value, present := out["finish_reason"]
		assert.True(t, present)
		assert.Equal(t, want, value)
		for _, format := range []struct {
			name   string
			encode func(any, ...*Options) ([]byte, error)
		}{
			{"json", UnbindJSON},
			{"jsonl", UnbindJSONL},
			{"yaml", UnbindYAML},
		} {
			t.Run(format.name, func(t *testing.T) {
				output, err := format.encode(target)
				require.NoError(t, err)
				var decoded map[string]any
				if format.name == "yaml" {
					require.NoError(t, yaml.Unmarshal(output, &decoded))
				} else {
					require.NoError(t, json.Unmarshal(output, &decoded))
				}
				if format.name == "jsonl" {
					assert.Equal(t, byte('\n'), output[len(output)-1])
				}
				value, present := decoded["finish_reason"]
				assert.True(t, present)
				assert.Equal(t, want, value)
			})
		}
	}
}
