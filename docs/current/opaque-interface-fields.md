# Opaque interface fields and nullable output

`dd` supports consumer-owned field shapes through `+opaque`, and named-field null output through `+nullable`. All JSON, JSONL, and YAML output helpers use the same `Unbind` pipeline.

```go
type Request struct {
    Stop any `dd:",+opaque,+nullable,+omitempty"`
}

type Message struct {
    Content any `dd:",+opaque,+nullable"`
}

type Chunk struct {
    FinishReason *string `dd:",+nullable"`
}
```

## Intake and assignment

`+opaque` accepts exactly `map[string]any` (objects only) or an empty interface, including a named one (any non-null value). It assigns raw input without coercion or copying, preserving concrete Go types and sharing input maps and slices. A supplied opaque object replaces the field wholesale during `Merge`; missing fields preserve their existing values. Opaque assignment bypasses converters, dynamic binders, and unmarshalers. Untagged interface binding, `Dynamic` resolution, and existing opaque tag precedence are unchanged.

Bare null requires `+nullable` and then behaves as absence, preserving an existing value in both `Bind` and `Merge`. Null or missing input with `+required` fails. Direct map binding carries typed nil values as concrete Go values rather than normalizing them to bare null.

Strict intake checks the entire document, including opaque values: duplicate keys and prohibited YAML syntax still fail, and numbers retain authored spellings as `json.Number`. Arbitrary keys within an opaque value succeed; unknown siblings fail unless collected by `+extra`. Direct map binding does not replay byte-level checks. Strict `Merge` remains unsupported.

## Null and omission

`+nullable` emits null for nil pointers, interfaces, maps, and slices. Interfaces holding typed nil pointers/maps/slices also emit null. Zero primitives remain primitive values. Non-nil pointers are not dereferenced by the null check; non-nil values follow the usual output conversion path, including custom hooks, `Dynamic`, and key collision checks. Unsupported kinds remain unsupported.

`+omitempty` takes precedence using the existing field-level emptiness test. A nil interface is omitted, but a non-nil interface holding a typed nil emits null. Non-nil pointers to zero values remain present, subject to the existing post-conversion omission of empty maps.

| Field state | Nullable output | With `+omitempty` |
| --- | --- | --- |
| Nil pointer or interface | Null | Omitted |
| Nil map or slice | Null | Omitted |
| Interface holding typed nil pointer/map/slice | Null | Null |
| Non-nil pointer to empty string | Empty string | Empty string |
| Zero int | Zero | Omitted |

The rule applies to named fields in nested structs and flattened value/non-nil pointer embeds. Nil embedded struct pointers contribute no fields. `+extra` contributes entries rather than a named null field. Untagged nil output and recursive map/slice element behavior are unchanged.

## Consumer behavior and migration

`{"stop":"END"}` and `{"stop":["END","DONE"]}` both bind and unbind with their string/list shapes. A zero `Chunk` emits `{"finish_reason":null}`; once populated with `"stop"`, it emits `{"finish_reason":"stop"}`. Missing and null input collapse to the same state, so a zero `Message` emits `{"content":null}` for either input.

Nullable nil output intentionally changes from omission for pointers/interfaces and from empty collections for maps/slices. Add `+omitempty` for omission under its field-level test. `+opaque` controls intake only and does not bypass output conversion.

## Boundaries

Presence tracking and null-as-clear updates are not represented by these tags. Consumers own typed union and allowed-shape validation. General interface binding, recursive tag validation, and unrelated collection binding are separate concerns. Direct `Bind` is not a JSON-value validator: opaque fields may carry arbitrary Go values, whose eventual serializability depends on existing conversion and encoder rules. Arbitrary object identity and original JSON/YAML bytes are not round-trip guarantees.
