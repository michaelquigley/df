package df

type nestedType struct {
	Name  string
	Count int
}

// Test helpers for Dynamic
type dynA struct{ Name string }

func (d *dynA) Type() string          { return "a" }
func (d *dynA) ToMap() map[string]any { return map[string]any{"type": "a", "name": d.Name} }

type dynB struct{ Count int }

func (d *dynB) Type() string          { return "b" }
func (d *dynB) ToMap() map[string]any { return map[string]any{"type": "b", "count": d.Count} }
