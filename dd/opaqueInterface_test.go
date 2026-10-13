package dd

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opaqueEmpty interface{}

type opaqueDocument struct {
	Value any `dd:",+opaque"`
}

func TestOpaqueInterfaceIntake(t *testing.T) {
	for _, strict := range []bool{false, true} {
		opt := &Options{Strict: strict}
		number := any(float64(5))
		if strict {
			number = json.Number("5.00")
		}
		for _, test := range []struct {
			name string
			json string
			yaml string
			want any
		}{
			{"string", `"END"`, "END", "END"},
			{"boolean", "true", "true", true},
			{"number", "5.00", "5.00", number},
			{"list", `["END","DONE"]`, "[END, DONE]", []any{"END", "DONE"}},
			{"object", `{"arbitrary":true}`, "{arbitrary: true}", map[string]any{"arbitrary": true}},
			{"nested list", `[{"n":5.00},null,["END"]]`, "[{n: 5.00}, null, [END]]",
				[]any{map[string]any{"n": number}, nil, []any{"END"}}},
		} {
			t.Run(test.name+modeName(strict), func(t *testing.T) {
				var fromJSON, fromYAML opaqueDocument
				require.NoError(t, BindJSON(&fromJSON, []byte(`{"value":`+test.json+"}"), opt))
				require.NoError(t, BindYAML(&fromYAML, []byte("value: "+test.yaml+"\n"), opt))
				assert.Equal(t, test.want, fromJSON.Value)
				assert.Equal(t, test.want, fromYAML.Value)
			})
		}
	}
}

func modeName(strict bool) string {
	if strict {
		return "/strict"
	}
	return "/forgiving"
}

func TestOpaqueInterfaceDeclarations(t *testing.T) {
	for _, strict := range []bool{false, true} {
		opt := &Options{Strict: strict}
		var named struct {
			Value opaqueEmpty `dd:",+opaque"`
		}
		require.NoError(t, Bind(&named, map[string]any{"value": int16(7)}, opt))
		assert.Equal(t, int16(7), named.Value)

		for _, test := range []struct {
			name string
			new  func() any
		}{
			{"method interface", func() any {
				return &struct {
					Value interface{ Method() } `dd:",+opaque"`
				}{}
			}},
			{"Dynamic", func() any {
				return &struct {
					Value Dynamic `dd:",+opaque"`
				}{}
			}},
			{"pointer interface", func() any {
				return &struct {
					Value *any `dd:",+opaque"`
				}{}
			}},
			{"slice", func() any {
				return &struct {
					Value []string `dd:",+opaque"`
				}{}
			}},
			{"named map", func() any {
				type namedMap map[string]any
				return &struct {
					Value namedMap `dd:",+opaque"`
				}{}
			}},
		} {
			t.Run(test.name+modeName(strict), func(t *testing.T) {
				for _, data := range []map[string]any{{}, {"value": "END"}} {
					var mismatch *TypeMismatchError
					require.ErrorAs(t, Bind(test.new(), data, opt), &mismatch)
					assert.Contains(t, mismatch.Expected, "+opaque")
				}
			})
		}
		var untagged struct{ Value any }
		var binding *BindingError
		err := Bind(&untagged, map[string]any{"value": "END"}, opt)
		assert.ErrorAs(t, err, &binding)
		assert.Contains(t, err.Error(), "interface fields are not supported")

		var object struct {
			Value map[string]any `dd:",+opaque"`
		}
		for _, raw := range []any{"END", []any{"END"}, nil} {
			var mismatch *TypeMismatchError
			assert.ErrorAs(t, Bind(&object, map[string]any{"value": raw}, opt), &mismatch)
		}
	}
}

func TestOpaqueInterfaceDirectAssignment(t *testing.T) {
	for _, strict := range []bool{false, true} {
		opt := &Options{Strict: strict}
		for _, raw := range []any{int8(7), uint32(8), float32(1.5), json.Number("5.00"), (*int)(nil), map[string]any(nil), []string(nil)} {
			var target opaqueDocument
			require.NoError(t, Bind(&target, map[string]any{"value": raw}, opt))
			assert.Equal(t, raw, target.Value)
			assert.Equal(t, reflect.TypeOf(raw), reflect.TypeOf(target.Value))
		}
		sourceMap := map[string]any{"old": true}
		sourceSlice := []string{"old"}
		var target opaqueDocument
		require.NoError(t, Bind(&target, map[string]any{"value": sourceMap}, opt))
		sourceMap["new"] = true
		assert.Equal(t, true, target.Value.(map[string]any)["new"])
		require.NoError(t, Bind(&target, map[string]any{"value": sourceSlice}, opt))
		sourceSlice[0] = "new"
		assert.Equal(t, "new", target.Value.([]string)[0])
	}
}

func TestOpaqueInterfaceNullAndMerge(t *testing.T) {
	type nullable struct {
		Value any `dd:"content,+opaque,+nullable"`
	}
	for _, strict := range []bool{false, true} {
		opt := &Options{Strict: strict}
		t.Run(modeName(strict), func(t *testing.T) {
			var plain opaqueDocument
			var mismatch *TypeMismatchError
			require.ErrorAs(t, BindJSON(&plain, []byte(`{"value":null}`), opt), &mismatch)
			assert.Equal(t, "opaqueDocument.Value", mismatch.Path)
			target := nullable{Value: "existing"}
			require.NoError(t, BindJSON(&target, []byte(`{"content":null}`), opt))
			assert.Equal(t, "existing", target.Value)
			require.NoError(t, BindJSON(&target, []byte(`{}`), opt))
			assert.Equal(t, "existing", target.Value)

			for _, data := range []map[string]any{{}, {"value": nil}} {
				var required struct {
					Value any `dd:",+opaque,+nullable,+required"`
				}
				var missing *RequiredFieldError
				assert.ErrorAs(t, Bind(&required, data, opt), &missing)
			}
		})
	}
	for _, test := range []struct {
		name string
		data map[string]any
		want any
	}{
		{"absent", map[string]any{}, map[string]any{"old": true}},
		{"null", map[string]any{"content": nil}, map[string]any{"old": true}},
		{"replace object", map[string]any{"content": map[string]any{"new": true}}, map[string]any{"new": true}},
		{"replace list", map[string]any{"content": []any{"new"}}, []any{"new"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := nullable{Value: map[string]any{"old": true}}
			require.NoError(t, Merge(&target, test.data))
			assert.Equal(t, test.want, target.Value)
		})
	}
	var target nullable
	var validation *ValidationError
	assert.ErrorAs(t, Merge(&target, map[string]any{"content": "new"}, Strict()), &validation)
}

func TestOpaqueInterfaceStrictBoundary(t *testing.T) {
	for _, data := range []string{
		`{"value":{"k":1,"k":2}}`,
		`{"value":[{"k":1,"k":2}]}`,
	} {
		var target opaqueDocument
		var duplicate *DuplicateKeyError
		assert.ErrorAs(t, BindJSON(&target, []byte(data), Strict()), &duplicate)
	}
	for _, data := range []string{
		"value: {k: 1, k: 2}\n",
		"value: [{k: 1, k: 2}]\n",
	} {
		var target opaqueDocument
		var duplicate *DuplicateKeyError
		assert.ErrorAs(t, BindYAML(&target, []byte(data), Strict()), &duplicate)
	}
	var target opaqueDocument
	require.NoError(t, BindJSON(&target, []byte(`{"value":{"unknown":{"arbitrary":true}}}`), Strict()))
	var unknown *UnknownFieldError
	require.ErrorAs(t, BindJSON(&target, []byte(`{"value":[],"unknown":true}`), Strict()), &unknown)
	assert.Equal(t, "unknown", unknown.Key)
	assert.Error(t, BindYAML(&target, []byte("value: [2026-10-13]\n"), Strict()))
}

type OpaqueEmbedded struct {
	Value any `dd:"content,+opaque,+nullable"`
}

func TestOpaqueInterfaceVisitedDeclarations(t *testing.T) {
	type InvalidOpaqueEmbedded struct {
		Value []string `dd:",+opaque"`
	}
	for _, strict := range []bool{false, true} {
		opt := &Options{Strict: strict}
		var target struct{ *InvalidOpaqueEmbedded }
		require.NoError(t, Bind(&target, map[string]any{}, opt))
		assert.Nil(t, target.InvalidOpaqueEmbedded)
		target.InvalidOpaqueEmbedded = &InvalidOpaqueEmbedded{}
		var mismatch *TypeMismatchError
		assert.ErrorAs(t, Bind(&target, map[string]any{}, opt), &mismatch)
		var valueEmbed struct{ InvalidOpaqueEmbedded }
		assert.ErrorAs(t, Bind(&valueEmbed, map[string]any{}, opt), &mismatch)
		var nested struct{ Nested InvalidOpaqueEmbedded }
		assert.ErrorAs(t, Bind(&nested, map[string]any{"nested": map[string]any{}}, opt), &mismatch)
	}
}

func TestOpaqueInterfaceTraversal(t *testing.T) {
	for _, strict := range []bool{false, true} {
		opt := &Options{Strict: strict}
		t.Run(modeName(strict), func(t *testing.T) {
			var target struct {
				Nested opaqueDocument
				Extra  map[string]any `dd:",+extra"`
			}
			require.NoError(t, BindJSON(&target, []byte(`{"nested":{"value":"END"},"extension":true}`), opt))
			assert.Equal(t, "END", target.Nested.Value)
			assert.Equal(t, map[string]any{"extension": true}, target.Extra)

			var embedded struct{ OpaqueEmbedded }
			require.NoError(t, BindJSON(&embedded, []byte(`{"content":"END"}`), opt))
			assert.Equal(t, "END", embedded.Value)
			var pointer struct{ *OpaqueEmbedded }
			require.NoError(t, BindJSON(&pointer, []byte(`{"content":null}`), opt))
			assert.Nil(t, pointer.OpaqueEmbedded)
			require.NoError(t, BindJSON(&pointer, []byte(`{"content":"END"}`), opt))
			require.NotNil(t, pointer.OpaqueEmbedded)
			assert.Equal(t, "END", pointer.Value)

			var extra struct {
				*OpaqueEmbedded
				Extra map[string]any `dd:",+extra"`
			}
			require.NoError(t, BindJSON(&extra, []byte(`{"content":null,"extension":true}`), opt))
			assert.Nil(t, extra.OpaqueEmbedded)
			assert.Equal(t, map[string]any{"extension": true}, extra.Extra)

			var skipped struct {
				Value  []string `dd:"-"`
				hidden []string `dd:",+opaque"`
			}
			require.NoError(t, Bind(&skipped, map[string]any{}, opt))
			var match struct {
				Value any `dd:",+opaque,+match=expected"`
			}
			require.NoError(t, Bind(&match, map[string]any{"value": "other"}, opt))
			assert.Equal(t, "other", match.Value)
		})
	}
}

type opaqueHook struct{ Called bool }

func (h *opaqueHook) UnmarshalDd(map[string]any) error {
	h.Called = true
	return errors.New("opaque intake called UnmarshalDd")
}

func (h *opaqueHook) Type() string { return "hook" }

func (h *opaqueHook) ToMap() (map[string]any, error) {
	return map[string]any{"type": "hook"}, nil
}

type opaqueConverter struct{ FromCalled, ToCalled bool }

func (c *opaqueConverter) FromRaw(any) (any, error) {
	c.FromCalled = true
	return nil, errors.New("opaque intake called FromRaw")
}

func (c *opaqueConverter) ToRaw(any) (any, error) {
	c.ToCalled = true
	return "converted", nil
}

func TestOpaqueInterfaceHooks(t *testing.T) {
	for _, strict := range []bool{false, true} {
		converter := &opaqueConverter{}
		binderCalled := false
		opt := &Options{
			Strict: strict,
			Converters: map[reflect.Type]Converter{
				reflect.TypeOf((*any)(nil)).Elem(): converter,
			},
			DynamicBinders: map[string]func(map[string]any) (Dynamic, error){
				"hook": func(map[string]any) (Dynamic, error) {
					binderCalled = true
					return nil, errors.New("opaque intake called Dynamic binder")
				},
			},
		}
		hook := &opaqueHook{}
		target := opaqueDocument{Value: hook}
		require.NoError(t, Bind(&target, map[string]any{"value": hook}, opt))
		assert.Same(t, hook, target.Value)
		assert.False(t, hook.Called)
		require.NoError(t, Bind(&target, map[string]any{"value": map[string]any{"type": "hook"}}, opt))
		assert.False(t, converter.FromCalled)
		assert.False(t, binderCalled)
		out, err := Unbind(target, opt)
		require.NoError(t, err)
		assert.True(t, converter.ToCalled)
		assert.Equal(t, "converted", out["value"])

		out, err = Unbind(opaqueDocument{Value: hook})
		require.NoError(t, err)
		assert.Equal(t, map[string]any{"type": "hook"}, out["value"])
	}
	var collision *KeyCollisionError
	_, err := Unbind(opaqueDocument{Value: map[any]string{1: "int", "1": "string"}})
	assert.ErrorAs(t, err, &collision)
}

func TestOpaqueInterfaceConsumerRoundTrip(t *testing.T) {
	type request struct {
		Stop any `dd:",+opaque,+nullable,+omitempty"`
	}
	type message struct {
		Content any `dd:",+opaque,+nullable"`
	}
	for _, strict := range []bool{false, true} {
		for _, data := range []string{`{"stop":"END"}`, `{"stop":["END","DONE"]}`} {
			var target request
			require.NoError(t, BindJSON(&target, []byte(data), &Options{Strict: strict}))
			output, err := UnbindJSON(target)
			require.NoError(t, err)
			assert.JSONEq(t, data, string(output))
		}
		for _, data := range []string{`{}`, `{"content":null}`, `{"content":"hello"}`, `{"content":["hello"]}`} {
			var target message
			require.NoError(t, BindJSON(&target, []byte(data), &Options{Strict: strict}))
			output, err := UnbindJSON(target)
			require.NoError(t, err)
			if data == "{}" {
				data = `{"content":null}`
			}
			assert.JSONEq(t, data, string(output))
		}
	}
}
