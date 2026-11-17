---
title: opaque accepts raw json fields
state: researching
created: 2026-11-17
tags: [enhancement]
subsystems: [dd]
source: mercurius
---

`+opaque` carries a JSON value through `dd` untouched on a `map[string]any` or `any` field, but it refuses a `json.RawMessage` field, and an untagged `json.RawMessage` unbinds as a list of byte values. let `+opaque` apply to `json.RawMessage` and `[]json.RawMessage`: on bind, the node at that key becomes the field's bytes; on unbind, the bytes are emitted as the nested value they encode, with number lexemes kept. a struct can then hold raw JSON as bytes and still be `dd`-bound, with no per-project converter.

## why

a struct that carries someone else's JSON — a reviewer's output, a tool call's arguments, a provider's response fragment — usually holds it as `json.RawMessage`, because the bytes are what gets validated, logged, and passed on. `dd` cannot bind that type, so each project that meets it invents a way around:

- terminus registers a `dd.Converter` for `json.RawMessage` (`internal/report/codec.go`) and routes every bind and unbind through one shared options value so disk and wire cannot diverge. its `ToRaw` decodes with a plain `json.Unmarshal` into `any`, so numbers pass through `float64`; an integer above 2^53 would not survive it.
- pane holds `[]json.RawMessage` in a `dd`-bound struct (`internal/llm/round.go`) and assembles the items by hand.
- mercurius changed its stored reviewer output from `json.RawMessage` to a `map[string]any` tagged `+opaque` to stay off a converter, and keeps the bytes separately where it needs them.

the `map[string]any` route works and is exact under strict binding. what is missing is the same treatment for a field that wants to stay bytes.

## discussion

verified against v1.0.6.

`+opaque` on a `json.RawMessage` field, strict bind of `{"name":"x","raw":{"n":1}}`:

```
R: expected map[string]any or empty interface for +opaque field, got jsontext.Value
```

an untagged `json.RawMessage` holding `{"n":9007199254740993}`, unbound:

```
{"name":"x","raw":[123,34,110,34,58,57,48,48,55,49,57,57,50,53,52,55,52,48,57,57,51,125]}
```

and binding an object into it:

```
binding field Plain.Raw from key "raw": Plain.Raw: expected array for slice, got map[string]interface {}
```

for comparison, the same document through `Raw map[string]any` tagged `+opaque` round-trips exactly under a strict bind (`{"f":1.50,"n":9007199254740993,...}` out), and loses digits under a forgiving one (`"f":1.5,"n":9007199254740992`), because forgiving intake decodes numbers through `float64`.

notes for whoever takes it:

- that last point decides something for this card. bytes built from a forgivingly-decoded node inherit the rounding. either the raw path keeps the number's lexeme in both modes, or exactness is documented as strict-only, as it is for the map form today.
- the error above names the type `jsontext.Value`: under the current toolchain `json.RawMessage` is an alias for it, so match on the type, not the name.
- unbind of a nil or empty raw field should follow the nil-pointer rule (omitted, or `null` under `+nullable`), not emit an empty value.
- the bytes produced by a bind are `dd`'s serialization of the node, not the bytes that were on the wire. say so in the doc comment; a caller that needs the original bytes has to keep them.
