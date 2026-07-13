package dd

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

type strictDoc struct {
	Name   string         `dd:"name"`
	Count  int            `dd:"count"`
	Rate   float64        `dd:"rate,+omitempty"`
	Active bool           `dd:"active,+omitempty"`
	Wait   time.Duration  `dd:"wait,+omitempty"`
	At     time.Time      `dd:"at,+omitempty"`
	Tags   []string       `dd:"tags,+omitempty"`
	Nested *strictNested  `dd:"nested,+omitempty"`
	Config map[string]any `dd:"config,+opaque"`
}

type strictNested struct {
	Value string `dd:"value"`
}

type strictReferenced struct{}

func (*strictReferenced) GetId() string { return "" }

type strictPointerDoc struct {
	Ref *Pointer[*strictReferenced] `dd:"ref"`
}

type strictPointerShapes struct {
	Direct *Pointer[*strictReferenced]            `dd:"direct"`
	Values []Pointer[*strictReferenced]           `dd:"values"`
	Slice  []*Pointer[*strictReferenced]          `dd:"slice"`
	Map    map[string]*Pointer[*strictReferenced] `dd:"map"`
}

type strictCustomName struct {
	Value string
}

func (n *strictCustomName) UnmarshalDd(data map[string]any) error {
	first, firstOK := data["first_name"].(string)
	last, lastOK := data["last_name"].(string)
	if !firstOK || !lastOK {
		return errors.New("first_name and last_name are required")
	}
	n.Value = first + " " + last
	return nil
}

func TestStrictJSONHappyPath(t *testing.T) {
	data := []byte(`{
		"name": "example",
		"count": 42,
		"rate": 0.5,
		"active": true,
		"wait": "30s",
		"at": "2026-07-10T12:00:00.123456789Z",
		"tags": ["a", "b"],
		"nested": {"value": "inner"},
		"config": {"anything": {"goes": 5.00}, "here": [1, 2]}
	}`)
	var doc strictDoc
	if err := BindJSON(&doc, data, Strict()); err != nil {
		t.Fatal(err)
	}
	if doc.Name != "example" || doc.Count != 42 || doc.Rate != 0.5 || !doc.Active {
		t.Fatalf("bound: %+v", doc)
	}
	if doc.Wait != 30*time.Second {
		t.Fatalf("wait: %v", doc.Wait)
	}
	if doc.At.Nanosecond() != 123456789 {
		t.Fatalf("RFC3339Nano precision lost: %v", doc.At)
	}
	if doc.Nested == nil || doc.Nested.Value != "inner" {
		t.Fatalf("nested: %+v", doc.Nested)
	}
	// the opaque subtree is captured raw, numbers as json.Number with the
	// authored lexeme
	inner, ok := doc.Config["anything"].(map[string]any)
	if !ok {
		t.Fatalf("config: %+v", doc.Config)
	}
	if n, ok := inner["goes"].(json.Number); !ok || string(n) != "5.00" {
		t.Fatalf("opaque number lexeme: %v (%T)", inner["goes"], inner["goes"])
	}
}

func TestStrictJSONDuplicateKeys(t *testing.T) {
	cases := []string{
		`{"name": "a", "name": "b", "count": 1}`,
		`{"name": "a", "count": 1, "nested": {"value": "x", "value": "y"}}`,
		// duplicate keys are rejected even inside the opaque subtree —
		// syntactic rules apply everywhere
		`{"name": "a", "count": 1, "config": {"k": 1, "k": 2}}`,
	}
	for _, doc := range cases {
		var target strictDoc
		err := BindJSON(&target, []byte(doc), Strict())
		var dup *DuplicateKeyError
		if !errors.As(err, &dup) {
			t.Errorf("%s: expected DuplicateKeyError, got %v", doc, err)
		}
	}
}

func TestStrictJSONTrailingData(t *testing.T) {
	var target strictDoc
	err := BindJSON(&target, []byte(`{"name": "a", "count": 1} extra`), Strict())
	if err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Fatalf("trailing data accepted or wrong error: %v", err)
	}
}

func TestStrictUnknownField(t *testing.T) {
	var target strictDoc
	err := BindJSON(&target, []byte(`{"name": "a", "count": 1, "surprise": true}`), Strict())
	var unknown *UnknownFieldError
	if !errors.As(err, &unknown) {
		t.Fatalf("expected UnknownFieldError, got %v", err)
	}
	if unknown.Key != "surprise" {
		t.Fatalf("unknown key: %q", unknown.Key)
	}

	// nested unknowns carry the nested path
	err = BindJSON(&target, []byte(`{"name": "a", "count": 1, "nested": {"value": "x", "extra": 1}}`), Strict())
	if !errors.As(err, &unknown) {
		t.Fatalf("expected nested UnknownFieldError, got %v", err)
	}
}

func TestStrictCoercionRefusals(t *testing.T) {
	cases := map[string]string{
		`{"name": 5, "count": 1}`:                     "expected string",  // number → string
		`{"name": "a", "count": "5"}`:                 "expected integer", // string → int
		`{"name": "a", "count": 5.5}`:                 "not an integer",   // fraction → int
		`{"name": "a", "count": 5e2}`:                 "not an integer",   // exponent → int
		`{"name": "a", "count": 1, "active": "true"}`: "expected bool",    // string → bool
		`{"name": "a", "count": 1, "wait": 30}`:       "duration string",  // number → duration
		`{"name": "a", "count": 1, "at": 1720620000}`: "time",             // number → time
	}
	for doc, wantErr := range cases {
		var target strictDoc
		err := BindJSON(&target, []byte(doc), Strict())
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("%s: expected error containing %q, got %v", doc, wantErr, err)
		}
	}
}

func TestStrictRejectsNativeTimeValue(t *testing.T) {
	var target strictDoc
	err := Bind(&target, map[string]any{
		"name":  "a",
		"count": 1,
		"at":    time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC),
	}, Strict())
	var mismatch *TypeMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("expected TypeMismatchError, got %v", err)
	}
}

func TestStrictNumericExactness(t *testing.T) {
	// integer lexemes bind to float fields (one JSON number type)
	var target strictDoc
	if err := BindJSON(&target, []byte(`{"name": "a", "count": 1, "rate": 5}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if target.Rate != 5 {
		t.Fatalf("rate: %v", target.Rate)
	}

	// overflow is refused, never wrapped
	type small struct {
		N int8 `dd:"n"`
	}
	var s small
	if err := BindJSON(&s, []byte(`{"n": 300}`), Strict()); err == nil || !strings.Contains(err.Error(), "overflows") {
		t.Fatalf("overflow accepted or wrong error: %v", err)
	}

	// negatives are refused for unsigned fields
	type unsigned struct {
		N uint `dd:"n"`
	}
	var u unsigned
	if err := BindJSON(&u, []byte(`{"n": -1}`), Strict()); err == nil || !strings.Contains(err.Error(), "negative") {
		t.Fatalf("negative-to-unsigned accepted or wrong error: %v", err)
	}

	// uint64 retains its full range rather than passing through int64
	type wideUnsigned struct {
		N uint64 `dd:"n"`
	}
	var wide wideUnsigned
	if err := BindJSON(&wide, []byte(`{"n": 18446744073709551615}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if wide.N != ^uint64(0) {
		t.Fatalf("uint64 max: %d", wide.N)
	}
	if err := BindJSON(&wide, []byte(`{"n": 18446744073709551616}`), Strict()); err == nil {
		t.Fatal("uint64 overflow accepted")
	}

	// native signed and unsigned values do not cross-bind in strict mode
	var nativeUnsigned unsigned
	if err := Bind(&nativeUnsigned, map[string]any{"n": int64(1)}, Strict()); err == nil {
		t.Fatal("native signed value bound to unsigned field")
	}
	type signed struct {
		N int64 `dd:"n"`
	}
	var nativeSigned signed
	if err := Bind(&nativeSigned, map[string]any{"n": uint64(1)}, Strict()); err == nil {
		t.Fatal("native unsigned value bound to signed field")
	}
}

func TestStrictMapKeys(t *testing.T) {
	type namedKey string
	type stringMaps struct {
		Plain map[string]int   `dd:"plain"`
		Named map[namedKey]int `dd:"named"`
	}
	var stringsOnly stringMaps
	if err := BindJSON(&stringsOnly, []byte(`{"plain":{"one":1},"named":{"two":2}}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if stringsOnly.Plain["one"] != 1 || stringsOnly.Named[namedKey("two")] != 2 {
		t.Fatalf("string-keyed maps: %+v", stringsOnly)
	}

	type intMap struct {
		Values map[int]string `dd:"values"`
	}
	for _, doc := range []string{
		`{"values":{"1":"one"}}`,
		`{"values":{}}`,
	} {
		var strict intMap
		err := BindJSON(&strict, []byte(doc), Strict())
		var mismatch *TypeMismatchError
		if !errors.As(err, &mismatch) {
			t.Fatalf("%s: expected TypeMismatchError, got %v", doc, err)
		}
	}

	var forgiving intMap
	if err := BindJSON(&forgiving, []byte(`{"values":{"1":"one"}}`)); err != nil {
		t.Fatal(err)
	}
	if forgiving.Values[1] != "one" {
		t.Fatalf("forgiving map-key conversion: %+v", forgiving.Values)
	}
}

func TestStrictExtraFieldCapturesByIntent(t *testing.T) {
	type withExtra struct {
		Name  string         `dd:"name"`
		Extra map[string]any `dd:",+extra"`
	}
	var target withExtra
	if err := BindJSON(&target, []byte(`{"name": "a", "anything": 1}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if _, ok := target.Extra["anything"]; !ok {
		t.Fatalf("extra: %+v", target.Extra)
	}
}

func TestStrictEmbeddedExtraField(t *testing.T) {
	type Extensions struct {
		Extra map[string]any `dd:",+extra"`
	}
	type valueDocument struct {
		Extensions
		Name string `dd:"name"`
	}
	var valueDoc valueDocument
	if err := BindJSON(&valueDoc, []byte(`{"name":"example","future_field":42}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if valueDoc.Name != "example" || len(valueDoc.Extra) != 1 {
		t.Fatalf("value-embedded extra: %+v", valueDoc)
	}
	if number, ok := valueDoc.Extra["future_field"].(json.Number); !ok || string(number) != "42" {
		t.Fatalf("future_field: %v (%T)", valueDoc.Extra["future_field"], valueDoc.Extra["future_field"])
	}
	if _, captured := valueDoc.Extra["name"]; captured {
		t.Fatalf("declared parent field captured as extra: %+v", valueDoc.Extra)
	}

	type pointerDocument struct {
		*Extensions
		Name string `dd:"name"`
	}
	var pointerDoc pointerDocument
	if err := BindJSON(&pointerDoc, []byte(`{"name":"example","future_field":42}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if pointerDoc.Extensions == nil || len(pointerDoc.Extra) != 1 {
		t.Fatalf("pointer-embedded extra: %+v", pointerDoc)
	}

	var noExtras pointerDocument
	if err := BindJSON(&noExtras, []byte(`{"name":"example"}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if noExtras.Extensions != nil {
		t.Fatalf("pointer embedding allocated without extras: %+v", noExtras)
	}
}

func TestStrictEmbeddedMultipleExtraFields(t *testing.T) {
	type ExtraOne struct {
		Extra map[string]any `dd:",+extra"`
	}
	type ExtraTwo struct {
		Extra map[string]any `dd:",+extra"`
	}
	type invalidDocument struct {
		ExtraOne
		ExtraTwo
	}
	var doc invalidDocument
	err := BindJSON(&doc, []byte(`{}`), Strict())
	var multiple *MultipleExtraFieldsError
	if !errors.As(err, &multiple) {
		t.Fatalf("expected MultipleExtraFieldsError, got %v", err)
	}
}

func TestStrictPointerUnknownFields(t *testing.T) {
	var strict strictPointerDoc
	if err := BindJSON(&strict, []byte(`{"ref":{"$ref":"target"}}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if strict.Ref == nil || strict.Ref.Ref != "target" {
		t.Fatalf("strict pointer: %+v", strict.Ref)
	}

	err := BindJSON(&strict, []byte(`{"ref":{"$ref":"target","zeta":1,"alpha":2}}`), Strict())
	var unknown *UnknownFieldError
	if !errors.As(err, &unknown) {
		t.Fatalf("expected UnknownFieldError, got %v", err)
	}
	if unknown.Key != "alpha" {
		t.Fatalf("unknown key: %q", unknown.Key)
	}

	var forgiving strictPointerDoc
	if err := BindJSON(&forgiving, []byte(`{"ref":{"$ref":"target","role":"admin"}}`)); err != nil {
		t.Fatal(err)
	}
	if forgiving.Ref == nil || forgiving.Ref.Ref != "target" {
		t.Fatalf("forgiving pointer: %+v", forgiving.Ref)
	}
}

func TestStrictPointerSpecializationAcrossShapes(t *testing.T) {
	valid := []byte(`{
		"direct":{"$ref":"direct"},
		"values":[{"$ref":"value"}],
		"slice":[{"$ref":"slice"}],
		"map":{"item":{"$ref":"map"}}
	}`)
	var shapes strictPointerShapes
	if err := BindJSON(&shapes, valid, Strict()); err != nil {
		t.Fatal(err)
	}
	if shapes.Direct.Ref != "direct" ||
		len(shapes.Values) != 1 || shapes.Values[0].Ref != "value" ||
		len(shapes.Slice) != 1 || shapes.Slice[0].Ref != "slice" ||
		shapes.Map["item"].Ref != "map" {
		t.Fatalf("pointer shapes: %+v", shapes)
	}

	cases := []string{
		`{"direct":{"$ref":"target","resolved":{}}}`,
		`{"values":[{"$ref":"target","resolved":{}}]}`,
		`{"slice":[{"$ref":"target","resolved":{}}]}`,
		`{"map":{"item":{"$ref":"target","resolved":{}}}}`,
	}
	for _, doc := range cases {
		var target strictPointerShapes
		err := BindJSON(&target, []byte(doc), Strict())
		var unknown *UnknownFieldError
		if !errors.As(err, &unknown) {
			t.Fatalf("%s: expected UnknownFieldError, got %v", doc, err)
		}
		if unknown.Key != "resolved" {
			t.Fatalf("%s: unknown key %q", doc, unknown.Key)
		}
	}
}

func TestStrictUnmarshalerOwnsCustomSchema(t *testing.T) {
	type document struct {
		Name strictCustomName `dd:"name"`
	}
	var doc document
	if err := BindJSON(&doc, []byte(`{"name":{"first_name":"Ada","last_name":"Lovelace"}}`), Strict()); err != nil {
		t.Fatal(err)
	}
	if doc.Name.Value != "Ada Lovelace" {
		t.Fatalf("custom name: %q", doc.Name.Value)
	}
}

func TestStrictRequiredStillEnforced(t *testing.T) {
	type withRequired struct {
		Name string `dd:"name,+required"`
	}
	var target withRequired
	err := BindJSON(&target, []byte(`{}`), Strict())
	var required *RequiredFieldError
	if !errors.As(err, &required) {
		t.Fatalf("expected RequiredFieldError, got %v", err)
	}
}

func TestStrictOpaqueRequiresObject(t *testing.T) {
	var target strictDoc
	err := BindJSON(&target, []byte(`{"name": "a", "count": 1, "config": "not an object"}`), Strict())
	if err == nil || !strings.Contains(err.Error(), "+opaque") {
		t.Fatalf("non-object opaque accepted or wrong error: %v", err)
	}

	// +opaque demands map[string]any at the field
	type badOpaque struct {
		Config string `dd:"config,+opaque"`
	}
	var bad badOpaque
	if err := BindJSON(&bad, []byte(`{"config": {}}`), Strict()); err == nil || !strings.Contains(err.Error(), "map[string]any") {
		t.Fatalf("mistyped opaque field accepted or wrong error: %v", err)
	}
}

func TestStrictYAMLLexemePreservation(t *testing.T) {
	data := []byte("name: a\ncount: 1\nconfig:\n  cost: 5.00\n  plain: 7\n")
	var target strictDoc
	if err := BindYAML(&target, data, Strict()); err != nil {
		t.Fatal(err)
	}
	if n, ok := target.Config["cost"].(json.Number); !ok || string(n) != "5.00" {
		t.Fatalf("authored 5.00 did not survive: %v (%T)", target.Config["cost"], target.Config["cost"])
	}
	if n, ok := target.Config["plain"].(json.Number); !ok || string(n) != "7" {
		t.Fatalf("integer lexeme: %v (%T)", target.Config["plain"], target.Config["plain"])
	}
}

func TestStrictYAMLRejections(t *testing.T) {
	cases := map[string]string{
		"name: a\ncount: 1\ncount: 2\n":       "duplicate key",  // duplicate mapping key
		"base: &b {value: x}\nnested: *b\n":   "anchor",         // anchor/alias graph
		"name: &n a\ncount: 1\n":              "anchor",         // unused value anchor
		"&field name: a\ncount: 1\n":          "anchor",         // unused mapping-key anchor
		"name: a\ncount: 0x10\n":              "JSON number",    // hexadecimal integer
		"name: a\ncount: +16\n":               "JSON number",    // explicit-plus integer
		"name: a\ncount: 1\nat: 2026-07-10\n": "scalar tag",     // unquoted timestamp
		"name: a\ncount: 1\n---\nname: b\n":   "multi-document", // second document
		"5: a\n":                              "plain strings",  // non-string key
	}
	for doc, wantErr := range cases {
		var target strictDoc
		err := BindYAML(&target, []byte(doc), Strict())
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Errorf("%q: expected error containing %q, got %v", doc, wantErr, err)
		}
	}
}

func TestStrictDecodeTreeThenBind(t *testing.T) {
	// the strict decoders are public so a tree can be normalized between
	// intake and binding — sterling's authored-money canonicalization shape
	tree, err := DecodeStrictYAML([]byte("name: a\ncount: 1\nrate: 2.50\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n, ok := tree["rate"].(json.Number); !ok || string(n) != "2.50" {
		t.Fatalf("tree lexeme: %v", tree["rate"])
	}
	var target strictDoc
	if err := Bind(&target, tree, Strict()); err != nil {
		t.Fatal(err)
	}
	if target.Rate != 2.5 {
		t.Fatalf("rate: %v", target.Rate)
	}
}

func TestStrictMergeRefused(t *testing.T) {
	var target strictDoc
	err := MergeJSON(&target, []byte(`{"name": "a"}`), Strict())
	if err == nil || !strings.Contains(err.Error(), "not supported for merge") {
		t.Fatalf("strict merge accepted or wrong error: %v", err)
	}
}

// the forgiving path's helpfulness is now the load-bearing difference between
// modes, so it is pinned as promised behavior: a future change must not
// quietly strictify the default posture either.
func TestForgivingPathPins(t *testing.T) {
	type doc struct {
		Name  string `dd:"name"`
		Count int    `dd:"count"`
	}

	// unknown keys are ignored
	var d doc
	if err := BindJSON(&d, []byte(`{"name": "a", "count": 1, "surprise": true}`)); err != nil {
		t.Fatalf("forgiving path rejected unknown key: %v", err)
	}

	// duplicate JSON keys resolve last-wins inside the parser
	d = doc{}
	if err := BindJSON(&d, []byte(`{"name": "a", "name": "b", "count": 1}`)); err != nil {
		t.Fatalf("forgiving path rejected duplicate key: %v", err)
	}
	if d.Name != "b" {
		t.Fatalf("duplicate resolution: %q", d.Name)
	}

	// forgiving YAML retains yaml.v3's existing duplicate-key rejection
	d = doc{}
	if err := BindYAML(&d, []byte("name: a\nname: b\ncount: 1\n")); err == nil {
		t.Fatal("forgiving YAML accepted duplicate key")
	}

	// cross-type coercion works
	d = doc{}
	if err := BindJSON(&d, []byte(`{"name": "a", "count": "5"}`)); err != nil {
		t.Fatalf("forgiving path refused string→int coercion: %v", err)
	}
	if d.Count != 5 {
		t.Fatalf("coerced count: %d", d.Count)
	}

	// floats truncate into int fields
	d = doc{}
	if err := BindJSON(&d, []byte(`{"name": "a", "count": 5.9}`)); err != nil {
		t.Fatalf("forgiving path refused float→int coercion: %v", err)
	}
	if d.Count != 5 {
		t.Fatalf("truncated count: %d", d.Count)
	}
}
