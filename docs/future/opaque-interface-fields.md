# Opaque interface fields and nullable output

External formats sometimes give one field several possible shapes, and sometimes require a key to be present even when its value is null. `dd` should express both through its existing field tags. Extend `+opaque` to empty interfaces, and give `+nullable` an output meaning that composes with that extension.

This design covers [binding into interface fields](roadmap/binding-into-interface-fields.md) and [nullable output](roadmap/nullable-emits-explicit-null-on-unbind.md). The implementation plan is [opaque-interface-fields-work-order.md](opaque-interface-fields-work-order.md).

## Field types and interpretation

The Go type states what a field can hold; `+opaque` tells the binder to carry its input without interpreting it.

```go
type Request struct {
    Stop       any            `dd:",+opaque,+nullable,+omitempty"`
    ToolChoice any            `dd:",+opaque,+nullable,+omitempty"`
    Parameters map[string]any `dd:",+opaque"`
}

type Message struct {
    Content any `dd:",+opaque,+nullable"`
}

type Chunk struct {
    FinishReason *string `dd:",+nullable"`
}
```

An opaque `any` accepts a non-null scalar, list, or object. The assigned value retains its concrete Go type and shares the input's maps and slices; the binder neither coerces nor copies the subtree. This is a field-level assignment, including during `Merge`: a supplied non-null object replaces the field rather than recursively merging its members. Missing fields retain their existing values.

Opaque `map[string]any` fields retain their object requirement. Empty interfaces, including named empty interfaces, gain support; interfaces with methods, pointers to interfaces, and typed slices do not become opaque field types. Untagged interface binding and `Dynamic` resolution keep their existing behavior. No new tag or option is introduced.

The same assignment contract applies to direct `Bind` input: this is not a new JSON-value validator for arbitrary Go values. Whether a carried Go value can later be serialized remains subject to the existing unbinding and encoder rules.

## Strict mode

An opaque field declares that its contents belong to the consumer. Strict binding recognizes the field but does not validate members inside its value. Strict JSON/YAML intake still applies throughout the document, including opaque values: duplicate keys and prohibited syntax remain errors, and numbers retain the decoder's representation. Unknown keys outside declared fields still fail strict binding. Direct map binding does not replay byte-level intake checks. `Merge` continues to reject strict mode.

## Null and omission

`+opaque` does not grant null acceptance. A bare null is rejected unless the field also has `+nullable`. `+nullable` retains its intake meaning: explicit null behaves as absence. It leaves existing values untouched in both `Bind` and `Merge`, and null combined with `+required` is still a required-field error.

On output, `+nullable` makes nil values emit an explicit null. This covers pointer, empty-interface, map, and slice fields; an interface holding a typed nil pointer, map, or slice also emits null. It does not turn zero strings, numbers, or structs into null. Unsupported kinds do not gain serialization support.

`+omitempty` retains its existing emptiness test and takes precedence: if that test says to omit a field, `+nullable` cannot put it back. A non-nil interface containing a typed nil is not a zero interface under that test and therefore emits null. A non-nil pointer to a zero value remains present.

| Field declaration | Nil field output |
| --- | --- |
| `Content any` with `+opaque` | omitted |
| `Content any` with `+opaque,+nullable` | `"content": null` |
| `Content any` with `+opaque,+nullable,+omitempty` | omitted |
| `FinishReason *string` with `+nullable` | `"finish_reason": null` |
| `FinishReason *string` with `+nullable,+omitempty` | omitted |

Without `+nullable`, existing output behavior remains unchanged. The new null rule is for named struct fields, including fields reached through nested and flattened structs. A nil embedded struct pointer still contributes no fields, and `+extra` still contributes entries rather than a named null field. No global change to map entries or slice elements is needed.

Missing input and explicit null collapse to the same field state. A zero `Message` therefore emits `{"content":null}` whether its input omitted content or explicitly supplied null. This is intentional; these tags do not record presence.

## Scenarios

For `Request.Stop`, `{"stop":"END"}` binds a string, and `{"stop":["END","DONE"]}` binds a list. Unbinding preserves that string-or-list shape. The application owns any validation of allowed shapes or list element types.

For streaming output, a `Chunk` with nil `FinishReason` emits `{"finish_reason":null}`. Once the pointer holds `"stop"`, it emits `{"finish_reason":"stop"}`. No format-specific marshaler is needed.

For partial updates, merging `{"content":null}` into a message whose content is `"hello"` preserves `"hello"`. Merging a new list replaces it. Null-as-clear is not introduced.

## Seam census

- **Binding / consumer validation:** carry opaque values; the consumer validates vendor-specific alternatives. Strict parsing remains responsible for syntax. Revisit if a consumer needs a typed union rather than raw data.
- **Binding / unbinding:** `+opaque` controls intake interpretation; non-null output retains the existing `dd` conversion path. `+nullable` controls named-field null emission. This preserves custom conversions and collision checks.
- **Model / transport:** express shape in Go types and `dd` tags; keep JSON/YAML framing in the existing I/O helpers.
- **Error by tier:** invalid opaque declarations and disallowed null input return binding errors with field context; the framework does not log or terminate the caller.

## Compatibility

Opaque map binding remains compatible. Nullable pointer and interface output intentionally changes from omitted to null; nullable nil maps and slices change from their existing collection output to null. Callers wanting omission can use `+omitempty`. Existing non-null conversions, `Dynamic` behavior, strict intake, and nullable input semantics remain intact.

## Deferred (and Why)

Presence tracking and null-as-clear updates need an explicit representation or operation; neither follows from `any` or `+nullable`. Typed union validation belongs to the consumer until a concrete need justifies a separate API. General interface binding, recursive tag validation, and repairs to unrelated collection binding are outside these two cards. Arbitrary Go object identity and byte-for-byte JSON/YAML preservation are not round-trip guarantees: the existing unbinder interprets concrete values and encoders choose formatting.
