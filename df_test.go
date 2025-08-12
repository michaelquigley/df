package df

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicBind(t *testing.T) {
	basic := &struct {
		StringValue string
	}{}

	data := map[string]any{
		"string_value": "oh, wow!",
	}

	err := Bind(basic, data)
	assert.Nil(t, err)
	assert.Equal(t, "oh, wow!", basic.StringValue)
}

func TestRenaming(t *testing.T) {
	renamed := &struct {
		SomeInt int `df:"some_int_,required"`
	}{}

	data := map[string]any{
		"some_int_": 46,
	}

	err := Bind(renamed, data)
	assert.Nil(t, err)
	assert.Equal(t, 46, renamed.SomeInt)
}

func TestStringArray(t *testing.T) {
	withArray := &struct {
		StringArray []string
	}{}

	data := map[string]any{
		"string_array": []string{"one", "two", "three"},
	}

	err := Bind(withArray, data)
	assert.Nil(t, err)
	assert.EqualValues(t, []string{"one", "two", "three"}, withArray.StringArray)
}

func TestIntArray(t *testing.T) {
	withArray := &struct {
		IntArray []int
	}{}

	data := map[string]any{
		"int_array": []int{1, 2, 3, 4, 5, 6},
	}

	err := Bind(withArray, data)
	assert.Nil(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, withArray.IntArray)
}

func TestRequired(t *testing.T) {
	required := &struct {
		Required int `df:",required"`
	}{}

	data := make(map[string]any)

	err := Bind(required, data)
	assert.NotNil(t, err)
}

type nestedType struct {
	Name  string
	Count int
}

func TestNestedPtr(t *testing.T) {
	root := &struct {
		Id     string
		Nested *nestedType
	}{}

	data := map[string]any{
		"id": "TestNested",
		"nested": map[string]any{
			"name": "Different",
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "TestNested", root.Id)
	assert.NotNil(t, root.Nested)
	assert.Equal(t, "Different", root.Nested.Name)
	assert.Equal(t, 0, root.Nested.Count) // default zero value
}

func TestNestedValue(t *testing.T) {
	root := &struct {
		Id     string
		Nested nestedType
	}{}

	data := map[string]any{
		"id": "TestNested",
		"nested": map[string]any{
			"name": "Different",
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "TestNested", root.Id)
	assert.Equal(t, "Different", root.Nested.Name)
	assert.Equal(t, 0, root.Nested.Count)
}

func TestStructTypeArray(t *testing.T) {
	root := &struct {
		Id      string
		Nesteds []*nestedType
	}{}

	data := map[string]any{
		"id": "StructTypeArray",
		"nesteds": []map[string]any{
			{"name": "a"},
			{"name": "b"},
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "StructTypeArray", root.Id)
	assert.Equal(t, 2, len(root.Nesteds))
	assert.Equal(t, "a", root.Nesteds[0].Name)
	assert.Equal(t, "b", root.Nesteds[1].Name)
}

func TestAnonymousStruct(t *testing.T) {
	root := &struct {
		Id     string
		Nested struct {
			Name string
		}
	}{}

	data := map[string]any{
		"id": "AnonymousStruct",
		"nested": map[string]any{
			"name": "oh, wow!",
		},
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "AnonymousStruct", root.Id)
	assert.Equal(t, "oh, wow!", root.Nested.Name)
}

func TestUint32(t *testing.T) {
	root := &struct {
		Id    string
		Count uint32
	}{}

	data := map[string]any{
		"id":    "uint32",
		"count": 33,
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, "uint32", root.Id)
	assert.Equal(t, uint32(33), root.Count)
}

func TestTimeDuration(t *testing.T) {
	root := &struct {
		Duration time.Duration
	}{}

	data := map[string]any{
		"duration": "30s",
	}

	err := Bind(root, data)
	assert.Nil(t, err)
	assert.Equal(t, time.Duration(30)*time.Second, root.Duration)
}

func TestFloatWithIntData(t *testing.T) {
	basic := &struct {
		FloatValue float64
	}{}

	data := map[string]any{
		"float_value": 56,
	}

	err := Bind(basic, data)
	assert.Nil(t, err)
	assert.Equal(t, float64(56.0), basic.FloatValue)
}

func TestSliceTypeCompatibility(t *testing.T) {
	// ensure we can accept []interface{} as input for a []string field
	withArray := &struct {
		Items []string
	}{}

	arr := []interface{}{"a", "b", "c"}
	data := map[string]any{
		"items": arr,
	}
	err := Bind(withArray, data)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, withArray.Items)
}

func TestSnakeCaseDefault(t *testing.T) {
	// Verify that untagged field names default to snake_case keys
	s := &struct{ OhWow int }{}
	data := map[string]any{"oh_wow": 42}
	err := Bind(s, data)
	assert.Nil(t, err)
	assert.Equal(t, 42, s.OhWow)
	// sanity: ensure reflect usage somewhere so reflect import is used
	_ = reflect.TypeOf(s).Elem().NumField()
}

func TestUnbindBasic(t *testing.T) {
	s := &struct{ StringValue string }{StringValue: "oh, wow!"}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"string_value": "oh, wow!"}, m)
}

func TestUnbindRenaming(t *testing.T) {
	s := &struct {
		SomeInt int `df:"some_int_,required"`
	}{SomeInt: 46}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"some_int_": 46}, m)
}

func TestUnbindArrays(t *testing.T) {
	s := &struct {
		StringArray []string
		IntArray    []int
	}{StringArray: []string{"one", "two", "three"}, IntArray: []int{1, 2, 3}}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"string_array": []interface{}{"one", "two", "three"},
		"int_array":    []interface{}{1, 2, 3},
	}, m)
}

func TestUnbindNestedPtrAndValue(t *testing.T) {
	s1 := &struct {
		Id     string
		Nested *nestedType
	}{Id: "NestedPtr", Nested: &nestedType{Name: "Different"}}
	m1, err := Unbind(s1)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "NestedPtr",
		"nested": map[string]any{
			"name":  "Different",
			"count": 0,
		},
	}, m1)

	s2 := &struct {
		Id     string
		Nested nestedType
	}{Id: "NestedVal", Nested: nestedType{Name: "V"}}
	m2, err := Unbind(s2)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "NestedVal",
		"nested": map[string]any{
			"name":  "V",
			"count": 0,
		},
	}, m2)
}

func TestUnbindStructTypeArray(t *testing.T) {
	s := &struct {
		Id      string
		Nesteds []*nestedType
	}{Id: "StructTypeArray", Nesteds: []*nestedType{{Name: "a"}, {Name: "b"}}}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "StructTypeArray",
		"nesteds": []interface{}{
			map[string]any{"name": "a", "count": 0},
			map[string]any{"name": "b", "count": 0},
		},
	}, m)
}

func TestUnbindAnonymousStruct(t *testing.T) {
	s := &struct {
		Id     string
		Nested struct{ Name string }
	}{Id: "AnonymousStruct"}
	s.Nested.Name = "oh, wow!"
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{
		"id": "AnonymousStruct",
		"nested": map[string]any{
			"name": "oh, wow!",
		},
	}, m)
}

func TestUnbindUint32(t *testing.T) {
	s := &struct {
		Id    string
		Count uint32
	}{Id: "uint32", Count: 33}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"id": "uint32", "count": uint32(33)}, m)
}

func TestUnbindTimeDuration(t *testing.T) {
	s := &struct{ Duration time.Duration }{Duration: 30 * time.Second}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"duration": "30s"}, m)
}

func TestUnbindSnakeCaseDefault(t *testing.T) {
	s := &struct{ OhWow int }{OhWow: 42}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"oh_wow": 42}, m)
}

func TestUnbindSkipTag(t *testing.T) {
	s := &struct {
		A int `df:"-"`
		B int
	}{A: 1, B: 2}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"b": 2}, m)
}

func TestUnbindOmitNilPtr(t *testing.T) {
	s := &struct {
		Name   string
		Nested *nestedType
	}{Name: "root", Nested: nil}
	m, err := Unbind(s)
	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"name": "root"}, m)
}

func TestUnbindDynamicSlice(t *testing.T) {
    type root struct {
        Items []Dynamic
    }

    r := &root{
        Items: []Dynamic{
            &dynA{Name: "alpha"},
            &dynB{Count: 42},
        },
    }

    m, err := Unbind(r)
    assert.NoError(t, err)
    assert.Equal(t, map[string]any{
        "items": []interface{}{
            map[string]any{"type": "a", "name": "alpha"},
            map[string]any{"type": "b", "count": 42},
        },
    }, m)
}

// Test helpers for Dynamic
type dynA struct{ Name string }
func (d *dynA) Type() string          { return "a" }
func (d *dynA) ToMap() map[string]any { return map[string]any{"type": "a", "name": d.Name} }

type dynB struct{ Count int }
func (d *dynB) Type() string          { return "b" }
func (d *dynB) ToMap() map[string]any { return map[string]any{"type": "b", "count": d.Count} }

func TestBindDynamicSlice(t *testing.T) {
    type root struct {
        Items []Dynamic
    }

    data := map[string]any{
        "items": []map[string]any{
            {"type": "a", "name": "alpha"},
            {"type": "b", "count": 42},
        },
    }

    opts := &Options{
        DynamicBinders: map[string]func(map[string]any) (Dynamic, error){
            "a": func(m map[string]any) (Dynamic, error) {
                name, _ := m["name"].(string)
                return &dynA{Name: name}, nil
            },
            "b": func(m map[string]any) (Dynamic, error) {
                countAny, _ := m["count"]
                count := 0
                switch v := countAny.(type) {
                case int:
                    count = v
                case int64:
                    count = int(v)
                case float64:
                    count = int(v)
                }
                return &dynB{Count: count}, nil
            },
        },
    }

    r := &root{}
    err := Bind(r, data, opts)
    assert.NoError(t, err)
    if assert.Len(t, r.Items, 2) {
        a, ok := r.Items[0].(*dynA)
        assert.True(t, ok)
        assert.Equal(t, "alpha", a.Name)

        b, ok := r.Items[1].(*dynB)
        assert.True(t, ok)
        assert.Equal(t, 42, b.Count)
    }
}

func TestBindDynamicPerFieldBinders(t *testing.T) {
    type root struct {
        Action Dynamic
        Metric Dynamic
    }

    data := map[string]any{
        "action": map[string]any{"type": "a", "name": "alpha"},
        "metric": map[string]any{"type": "b", "count": 7},
    }

    opts := &Options{
        FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
            "root.Action": {
                "a": func(m map[string]any) (Dynamic, error) {
                    name, _ := m["name"].(string)
                    return &dynA{Name: name}, nil
                },
            },
            "root.Metric": {
                "b": func(m map[string]any) (Dynamic, error) {
                    countAny, _ := m["count"]
                    count := 0
                    switch v := countAny.(type) {
                    case int:
                        count = v
                    case int64:
                        count = int(v)
                    case float64:
                        count = int(v)
                    }
                    return &dynB{Count: count}, nil
                },
            },
        },
    }

    r := &root{}
    err := Bind(r, data, opts)
    assert.NoError(t, err)
    if assert.NotNil(t, r.Action) {
        a, ok := r.Action.(*dynA)
        assert.True(t, ok)
        assert.Equal(t, "alpha", a.Name)
    }
    if assert.NotNil(t, r.Metric) {
        b, ok := r.Metric.(*dynB)
        assert.True(t, ok)
        assert.Equal(t, 7, b.Count)
    }
}

func TestBindDynamicPerFieldSliceBinders(t *testing.T) {
    type root struct {
        Items  []Dynamic
        Others []Dynamic
    }

    data := map[string]any{
        "items": []map[string]any{{"type": "a", "name": "A1"}, {"type": "a", "name": "A2"}},
        "others": []map[string]any{{"type": "b", "count": 1}, {"type": "b", "count": 2}},
    }

    opts := &Options{
        FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
            "root.Items": {
                "a": func(m map[string]any) (Dynamic, error) {
                    name, _ := m["name"].(string)
                    return &dynA{Name: name}, nil
                },
            },
            "root.Others": {
                "b": func(m map[string]any) (Dynamic, error) {
                    countAny, _ := m["count"]
                    count := 0
                    switch v := countAny.(type) {
                    case int:
                        count = v
                    case int64:
                        count = int(v)
                    case float64:
                        count = int(v)
                    }
                    return &dynB{Count: count}, nil
                },
            },
        },
    }

    r := &root{}
    err := Bind(r, data, opts)
    assert.NoError(t, err)
    if assert.Len(t, r.Items, 2) {
        a1, _ := r.Items[0].(*dynA)
        a2, _ := r.Items[1].(*dynA)
        assert.Equal(t, "A1", a1.Name)
        assert.Equal(t, "A2", a2.Name)
    }
    if assert.Len(t, r.Others, 2) {
        b1, _ := r.Others[0].(*dynB)
        b2, _ := r.Others[1].(*dynB)
        assert.Equal(t, 1, b1.Count)
        assert.Equal(t, 2, b2.Count)
    }
}

func TestUnbindThenBindWithPerFieldBinders(t *testing.T) {
    type root struct {
        Action Dynamic
        Others []Dynamic
    }

    // Start with concrete dynamic values
    r1 := &root{
        Action: &dynA{Name: "alpha"},
        Others: []Dynamic{&dynB{Count: 1}, &dynB{Count: 2}},
    }

    // Unbind to map
    m, err := Unbind(r1)
    assert.NoError(t, err)

    // Bind back using per-field dynamic binders
    opts := &Options{
        FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
            "root.Action": {
                "a": func(m map[string]any) (Dynamic, error) {
                    name, _ := m["name"].(string)
                    return &dynA{Name: name}, nil
                },
            },
            "root.Others": {
                "b": func(m map[string]any) (Dynamic, error) {
                    countAny, _ := m["count"]
                    count := 0
                    switch v := countAny.(type) {
                    case int:
                        count = v
                    case int64:
                        count = int(v)
                    case float64:
                        count = int(v)
                    }
                    return &dynB{Count: count}, nil
                },
            },
        },
    }

    r2 := &root{}
    err = Bind(r2, m, opts)
    assert.NoError(t, err)

    if assert.NotNil(t, r2.Action) {
        a, ok := r2.Action.(*dynA)
        assert.True(t, ok)
        assert.Equal(t, "alpha", a.Name)
    }
    if assert.Len(t, r2.Others, 2) {
        b1, _ := r2.Others[0].(*dynB)
        b2, _ := r2.Others[1].(*dynB)
        assert.Equal(t, 1, b1.Count)
        assert.Equal(t, 2, b2.Count)
    }
}

func TestUnbindConcreteThenBindWithPerFieldBinders(t *testing.T) {
    type rootConcrete struct {
        DoA   dynA
        Bs    []*dynB
    }
    type rootIface struct {
        DoA Dynamic
        Bs  []Dynamic
    }

    // Start with concrete types implementing Dynamic
    r1 := &rootConcrete{
        DoA: dynA{Name: "alpha"},
        Bs:  []*dynB{{Count: 1}, {Count: 2}},
    }

    m, err := Unbind(r1)
    assert.NoError(t, err)

    // Bind into interface-typed fields using per-field binders
    opts := &Options{
        FieldDynamicBinders: map[string]map[string]func(map[string]any) (Dynamic, error){
            "rootIface.DoA": {
                "a": func(m map[string]any) (Dynamic, error) {
                    name, _ := m["name"].(string)
                    return &dynA{Name: name}, nil
                },
            },
            "rootIface.Bs": {
                "b": func(m map[string]any) (Dynamic, error) {
                    countAny, _ := m["count"]
                    count := 0
                    switch v := countAny.(type) {
                    case int:
                        count = v
                    case int64:
                        count = int(v)
                    case float64:
                        count = int(v)
                    }
                    return &dynB{Count: count}, nil
                },
            },
        },
    }

    r2 := &rootIface{}
    err = Bind(r2, m, opts)
    assert.NoError(t, err)

    if assert.NotNil(t, r2.DoA) {
        a, ok := r2.DoA.(*dynA)
        assert.True(t, ok)
        assert.Equal(t, "alpha", a.Name)
    }
    if assert.Len(t, r2.Bs, 2) {
        b1, _ := r2.Bs[0].(*dynB)
        b2, _ := r2.Bs[1].(*dynB)
        assert.Equal(t, 1, b1.Count)
        assert.Equal(t, 2, b2.Count)
    }
}
