package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/michaelquigley/df/dd"
)

// Payload demonstrates a document whose exact spelling is the contract — the
// shape strict mode exists for (signed payloads, hash-pinned documents,
// normative wire formats).
type Payload struct {
	Kind   string         `dd:"kind,+required"`
	Count  int            `dd:"count"`
	Cost   string         `dd:"cost"`           // money as an exact decimal string
	Config map[string]any `dd:"config,+opaque"` // carried, never interpreted
}

func main() {
	fmt.Println("=== 1. strict happy path ===")
	good := []byte(`{"kind": "example", "count": 42, "cost": "5.00", "config": {"anything": 5.00}}`)
	var p Payload
	if err := dd.BindJSON(&p, good, dd.Strict()); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("bound: %+v\n", p)
	// the opaque subtree preserves the authored number lexeme
	fmt.Printf("opaque lexeme: %v (%T)\n\n", p.Config["anything"], p.Config["anything"])

	fmt.Println("=== 2. duplicate keys are rejected ===")
	dup := []byte(`{"kind": "example", "count": 1, "count": 2}`)
	var forgiving Payload
	_ = dd.BindJSON(&forgiving, dup) // forgiving path: last wins, silently
	fmt.Printf("forgiving: count=%d (last wins, silently)\n", forgiving.Count)
	if err := dd.BindJSON(&p, dup, dd.Strict()); err != nil {
		fmt.Printf("strict:    %v\n\n", err)
	}

	fmt.Println("=== 3. unknown fields are rejected ===")
	unknown := []byte(`{"kind": "example", "count": 1, "surprise": true}`)
	if err := dd.BindJSON(&p, unknown, dd.Strict()); err != nil {
		fmt.Printf("strict: %v\n\n", err)
	}

	fmt.Println("=== 4. type coercion is refused ===")
	coerced := []byte(`{"kind": "example", "count": "5"}`)
	var c Payload
	_ = dd.BindJSON(&c, coerced) // forgiving path: "5" becomes 5
	fmt.Printf("forgiving: count=%d (string coerced)\n", c.Count)
	if err := dd.BindJSON(&p, coerced, dd.Strict()); err != nil {
		fmt.Printf("strict:    %v\n\n", err)
	}

	fmt.Println("=== 5. money survives as a number never touches float ===")
	yaml := []byte("kind: example\ncount: 1\ncost: \"5.00\"\nconfig:\n  budget: 5.00\n")
	var y Payload
	if err := dd.BindYAML(&y, yaml, dd.Strict()); err != nil {
		log.Fatal(err)
	}
	if n, ok := y.Config["budget"].(json.Number); ok {
		fmt.Printf("authored YAML 5.00 arrives as json.Number %q — not float64(5)\n\n", n)
	}

	fmt.Println("=== 6. decode, normalize, then bind ===")
	// the strict decoders are public for pipelines that adjust the tree
	// between intake and binding — for example, canonicalizing an authored
	// bare number into the exact string form the contract demands.
	authored := []byte("kind: example\ncount: 1\ncost: 5.00\n")
	tree, err := dd.DecodeStrictYAML(authored)
	if err != nil {
		log.Fatal(err)
	}
	if n, ok := tree["cost"].(json.Number); ok {
		tree["cost"] = string(n) // canonicalize: number lexeme → exact string
	}
	var normalized Payload
	if err := dd.Bind(&normalized, tree, dd.Strict()); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("canonicalized cost: %q\n", normalized.Cost)
}
