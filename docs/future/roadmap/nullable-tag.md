---
title: nullable tag
state: evaluating
created: 2026-08-10
tags: [feature]
subsystems: [dd]
source: llm-gateway
---

a `+nullable` struct tag, per field, where an explicit null binds as absent and leaves the zero value. today `allowed_models: null` fails with `expected array for slice, got <nil>`. it composes with `+required` the obvious way — a `+required +nullable` field given an explicit null still fails as required-missing — and reads naturally beside `+required` as the other half of the presence vocabulary.

a global `Options` flag would be the wrong shape: it makes nullability a property of the document rather than of the field, and contracts are not written that way.

## why

lowest priority of the four llm-gateway requests, and the one with a clean workaround: `DecodeStrictYAML` → strip the nullable keys → `dd.Bind` is documented as the seam for exactly this, is about fifteen lines, and is arguably more faithful than a library flag since the carve-out is named fields rather than a global posture. the case for building it anyway is that serializers emit `null` for empty collections and unset optionals as a matter of course, so every strict contract consuming machine-written documents meets this immediately and re-derives the same fifteen lines.

## discussion

the two failures, as they land:

```
allowed_models: null  →  expected array for slice, got <nil>
expires_at: null      →  expected time (RFC3339 string), got <nil>
```

surfaced while llm-gateway adopted `dd.Strict()` on a `v1.0.1` → `v1.0.2` bump. in its published key-record contract three fields are explicitly nullable — a null `allowed_models` or `allowed_routes` means "no restriction," and `expires_at` is nullable outright — while `name`, `key`, `version`, and `keys` must keep rejecting a null. that per-field split is the argument against a global posture.

the workaround, verified working at about fifteen lines:

```go
m, err := dd.DecodeStrictYAML(data)   // syntax rules: dup keys, aliases, multi-doc
stripNulls(m, nullableFields)         // delete the three nullable keys when nil
err = dd.Bind(&doc, m, dd.Strict())   // binding rules: unknown fields, zero coercion
```

`DecodeStrictYAML`/`DecodeStrictJSON` being public is documented as the seam for exactly this, which is why this request ranked last of the four.
