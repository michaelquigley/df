package dd

import (
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
