package dd

import (
	"errors"
	"math"
	"testing"
)

// detItem is a nested struct used to verify that nested keys sort independently.
type detItem struct {
	Title string `dd:"title"`
	Rank  int    `dd:"rank"`
}

// detConfig exercises the full ordering surface of deterministic output:
// scalar fields declared out of alphabetical order, a slice (order preserved),
// a nested struct (keys sorted), a map with keys inserted out of order (keys
// sorted), and an +extra map whose keys interleave with the top-level fields.
type detConfig struct {
	Zebra  string         `dd:"zebra"`
	Apple  string         `dd:"apple"`
	Mango  string         `dd:"mango"`
	Tags   []string       `dd:"tags"`
	Nested detItem        `dd:"nested"`
	Scores map[string]int `dd:"scores"`
	Extra  map[string]any `dd:",+extra"`
}

func newDetConfig() detConfig {
	return detConfig{
		Zebra:  "z",
		Apple:  "a",
		Mango:  "m",
		Tags:   []string{"red", "green", "blue"},
		Nested: detItem{Title: "hello", Rank: 7},
		Scores: map[string]int{"gamma": 3, "alpha": 1, "beta": 2},
		Extra:  map[string]any{"omega": "last", "alpha_extra": "first"},
	}
}

// golden output. note: top-level fields are alphabetized (declaration order is
// not preserved), map keys are sorted, +extra keys (alpha_extra, omega) are
// interleaved with the rest rather than appended, nested keys (rank, title) sort
// independently, and slice element order (red, green, blue) is preserved as-is.
const detGoldenJSON = `{
  "alpha_extra": "first",
  "apple": "a",
  "mango": "m",
  "nested": {
    "rank": 7,
    "title": "hello"
  },
  "omega": "last",
  "scores": {
    "alpha": 1,
    "beta": 2,
    "gamma": 3
  },
  "tags": [
    "red",
    "green",
    "blue"
  ],
  "zebra": "z"
}`

const detGoldenYAML = `alpha_extra: first
apple: a
mango: m
nested:
    rank: 7
    title: hello
omega: last
scores:
    alpha: 1
    beta: 2
    gamma: 3
tags:
    - red
    - green
    - blue
zebra: z
`

// compact form of detGoldenJSON: the same keys in the same sorted order on one
// line, newline-terminated.
const detGoldenJSONL = `{"alpha_extra":"first","apple":"a","mango":"m","nested":{"rank":7,"title":"hello"},"omega":"last","scores":{"alpha":1,"beta":2,"gamma":3},"tags":["red","green","blue"],"zebra":"z"}` + "\n"

// TestUnbindJSONGolden locks the exact sorted-key JSON output. it fails if the
// serialized format ever drifts from the documented determinism guarantee.
func TestUnbindJSONGolden(t *testing.T) {
	data, err := UnbindJSON(newDetConfig())
	if err != nil {
		t.Fatalf("UnbindJSON failed: %v", err)
	}
	if string(data) != detGoldenJSON {
		t.Errorf("JSON output does not match golden.\n--- got ---\n%s\n--- want ---\n%s", data, detGoldenJSON)
	}
}

// TestUnbindJSONLGolden locks the exact compact sorted-key JSONL output.
func TestUnbindJSONLGolden(t *testing.T) {
	data, err := UnbindJSONL(newDetConfig())
	if err != nil {
		t.Fatalf("UnbindJSONL failed: %v", err)
	}
	if string(data) != detGoldenJSONL {
		t.Errorf("JSONL output does not match golden.\n--- got ---\n%s\n--- want ---\n%s", data, detGoldenJSONL)
	}
}

// TestUnbindYAMLGolden locks the exact sorted-key YAML output.
func TestUnbindYAMLGolden(t *testing.T) {
	data, err := UnbindYAML(newDetConfig())
	if err != nil {
		t.Fatalf("UnbindYAML failed: %v", err)
	}
	if string(data) != detGoldenYAML {
		t.Errorf("YAML output does not match golden.\n--- got ---\n%s\n--- want ---\n%s", data, detGoldenYAML)
	}
}

// TestUnbindDeterminism proves output is byte-stable across repeated calls. go
// randomizes map iteration on every range, so reusing the same value across many
// iterations defends specifically against that randomization leaking into output.
func TestUnbindDeterminism(t *testing.T) {
	const iterations = 100
	src := newDetConfig()

	firstJSON, err := UnbindJSON(src)
	if err != nil {
		t.Fatalf("UnbindJSON failed: %v", err)
	}
	firstYAML, err := UnbindYAML(src)
	if err != nil {
		t.Fatalf("UnbindYAML failed: %v", err)
	}
	firstJSONL, err := UnbindJSONL(src)
	if err != nil {
		t.Fatalf("UnbindJSONL failed: %v", err)
	}

	for i := 0; i < iterations; i++ {
		gotJSON, err := UnbindJSON(src)
		if err != nil {
			t.Fatalf("UnbindJSON failed on iteration %d: %v", i, err)
		}
		if string(gotJSON) != string(firstJSON) {
			t.Fatalf("JSON output varied on iteration %d.\n--- got ---\n%s\n--- first ---\n%s", i, gotJSON, firstJSON)
		}

		gotYAML, err := UnbindYAML(src)
		if err != nil {
			t.Fatalf("UnbindYAML failed on iteration %d: %v", i, err)
		}
		if string(gotYAML) != string(firstYAML) {
			t.Fatalf("YAML output varied on iteration %d.\n--- got ---\n%s\n--- first ---\n%s", i, gotYAML, firstYAML)
		}

		gotJSONL, err := UnbindJSONL(src)
		if err != nil {
			t.Fatalf("UnbindJSONL failed on iteration %d: %v", i, err)
		}
		if string(gotJSONL) != string(firstJSONL) {
			t.Fatalf("JSONL output varied on iteration %d.\n--- got ---\n%s\n--- first ---\n%s", i, gotJSONL, firstJSONL)
		}
	}
}

// collisionHolder carries the one input that can falsify the guarantee: an
// interface-keyed map whose distinct keys stringify alike.
type collisionHolder struct {
	M map[any]string `dd:"m"`
}

// TestUnbindRejectsCollidingMapKeys pins that a map whose distinct keys
// serialize to one spelling is refused rather than collapsed. last-write-wins
// would let go's randomized iteration pick the survivor, so unbind refuses,
// and the refusal itself reads the same on every run.
func TestUnbindRejectsCollidingMapKeys(t *testing.T) {
	src := collisionHolder{M: map[any]string{1: "int", "1": "str"}}

	_, err := Unbind(src)
	if err == nil {
		t.Fatal("expected Unbind to refuse colliding map keys")
	}
	var collision *KeyCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("expected *KeyCollisionError, got %T: %v", err, err)
	}
	if collision.Key != "1" {
		t.Errorf("expected colliding spelling %q, got %q", "1", collision.Key)
	}
	want := `unbinding field collisionHolder.M to key "m": map key collision: "1" (string), 1 (int) serialize to the same key "1"`
	if err.Error() != want {
		t.Errorf("error text mismatch.\n--- got ---\n%s\n--- want ---\n%s", err.Error(), want)
	}

	// the refusal must not depend on iteration order either: same text every time
	const iterations = 100
	for i := 0; i < iterations; i++ {
		_, err := Unbind(src)
		if err == nil {
			t.Fatalf("Unbind accepted colliding keys on iteration %d", i)
		}
		if err.Error() != want {
			t.Fatalf("error text varied on iteration %d.\n--- got ---\n%s\n--- want ---\n%s", i, err.Error(), want)
		}
	}

	// the serialized forms inherit the refusal
	if _, err := UnbindJSON(src); err == nil {
		t.Error("expected UnbindJSON to refuse colliding map keys")
	}
	if _, err := UnbindJSONL(src); err == nil {
		t.Error("expected UnbindJSONL to refuse colliding map keys")
	}
	if _, err := UnbindYAML(src); err == nil {
		t.Error("expected UnbindYAML to refuse colliding map keys")
	}
}

// TestUnbindInterfaceKeyedMap pins that the collision check bites only on a real
// collision: an interface-keyed map whose keys stringify distinctly unbinds as
// before.
func TestUnbindInterfaceKeyedMap(t *testing.T) {
	src := collisionHolder{M: map[any]string{1: "int", "b": "str"}}
	out, err := Unbind(src)
	if err != nil {
		t.Fatalf("Unbind failed: %v", err)
	}
	m, ok := out["m"].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any for m, got %T", out["m"])
	}
	if len(m) != 2 || m["1"] != "int" || m["b"] != "str" {
		t.Errorf("unexpected map contents: %v", m)
	}
}

// TestUnbindRejectsNaNMapKeys shows why the check is not limited to
// interface-keyed maps: NaN never equals itself, so a float-keyed map can hold
// several NaN entries, and every one of them stringifies to "NaN".
func TestUnbindRejectsNaNMapKeys(t *testing.T) {
	type holder struct {
		M map[float64]int `dd:"m"`
	}
	src := holder{M: map[float64]int{math.NaN(): 1, math.NaN(): 2}}
	if len(src.M) != 2 {
		t.Fatalf("expected two NaN entries, got %d", len(src.M))
	}
	_, err := Unbind(src)
	var collision *KeyCollisionError
	if !errors.As(err, &collision) {
		t.Fatalf("expected *KeyCollisionError, got %T: %v", err, err)
	}
	if collision.Key != "NaN" {
		t.Errorf("expected colliding spelling %q, got %q", "NaN", collision.Key)
	}
}
