---
title: binding into interface fields
state: evaluating
created: 2026-10-13
tags: [feature]
subsystems: [dd]
milestone: v1.0.x
source: llm-gateway
---

`dd` refuses to bind anything into an `any` field, so a struct modelling an external format that legitimately accepts two shapes in one place cannot be `dd`-bound at all. `Unbind` already handles interface fields; only binding refuses. Give a field an opt-in way to accept whatever node the input carries — scalar, list, or object — and set it verbatim, the way `encoding/json` fills an `any`.

## why

this is what currently keeps a whole class of wire types outside `dd`. llm-gateway models the OpenAI chat API, where four fields are polymorphic by the vendor's design: a message's `content` is a string or a list of parts, `stop` is a string or a list, `tool_choice` is a string or an object, and a tool's `parameters` is an arbitrary JSON schema. those shapes are not ours to normalize, so the package stays on `json` tags and `encoding/json` — which the `golang-conventions/df-binding` quality then flags as a parallel binding substrate on every new field. the gateway is complying by converting, and this is the gap that blocks it.

`+opaque` is the closest existing tag and does not fit: it requires an object, and these fields can be strings.

## discussion

the refusal, verified against v1.0.5:

```
request.Stop: interface fields are not supported     (dd/bind.go:886)
```

`setNonPtrValue`'s `reflect.Interface` case special-cases `Dynamic` and returns that error for every other interface type. the shape worth considering is a tag — `+any`, or widening `+opaque` past its object requirement — so the capability is declared by the struct rather than inferred from the field type, keeping an accidental `any` an error rather than a silent passthrough. the inverse direction already works, so a tagged field round-trips once binding lands.

worth deciding at the same time: whether a tagged interface field participates in strict mode, where "accept any node" and "refuse what you do not recognize" meet. the honest reading is that an interface field's contents are uninterpreted either way, which is the same position `+opaque` already takes inside strict intake.
