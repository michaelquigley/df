---
title: separable nullable intake and emission
state: inbox
created: 2026-10-13
tags: [enhancement]
subsystems: [dd]
source: llm-gateway
---

`+nullable` welds two independent decisions together: accepting an explicit null on intake, and emitting an explicit null for a nil value. A struct modelling an external format often needs the first without the second, and no tag combination provides it. Separate them — a second tag, or an intake-only spelling — so null tolerance and null emission can be chosen per field.

## why

the combinations available today, verified against v1.0.6 for a pointer field:

| tag | binds null | nil emits | empty non-nil emits |
| --- | --- | --- | --- |
| untagged | fails | omitted | `{}` |
| `+nullable` | ok | `null` | `{}` |
| `+nullable,+omitempty` | ok | omitted | omitted |
| `+omitempty` | fails | omitted | omitted |

the missing row is *accepts null, omits when nil, emits `{}` when present and empty* — which is exactly `encoding/json`'s behavior for a `,omitempty` pointer, and therefore what any struct migrating from `encoding/json` needs. llm-gateway hits it on the OpenAI streaming `delta`: the field must tolerate a null from a backend, vanish when the gateway has no delta, and serialize as `{}` on the terminal chunk where clients expect an empty object. one field, no reachable tag.

the same split would simplify the wider migration there: every field a peer may send as null needs intake tolerance, and only two fields in the whole wire model should ever *emit* null.

## discussion

`+nullable` reads naturally as the intake rule — that is what v1.0.3 introduced it for — and the emission half arrived later in v1.0.6. one shape is to keep `+nullable` as intake-only and give emission its own tag; another is a modifier that suppresses emission while keeping intake. either way the two halves become independently selectable, and `+omitempty` keeps its current meaning.

worth deciding at the same time whether intake tolerance should simply be the default for every field. `encoding/json` accepts a null into any field and leaves the zero value, and `dd` rejects it on everything untagged — a null into a plain `string` or a slice fails in both strict and forgiving mode. that difference is what forces a tag audit on every field of a migrating struct rather than just the pointers, and making tolerance the default would remove the audit entirely. it is also a behavior change to existing consumers, so it is a release consideration rather than a quiet fix.
