---
title: nullable emits explicit null on unbind
state: evaluating
created: 2026-10-13
tags: [enhancement]
subsystems: [dd]
source: llm-gateway
milestone: v1.0.x
---

`Unbind` omits a nil pointer field entirely, so a struct cannot emit an explicit `null` — and some external contracts require the key present and null rather than absent. make `+nullable` symmetric: it already reads an explicit null as absent, and it should write a nil pointer as an explicit null.

## why

v1.0.3's `+nullable` solved the read side of this. the write side is still unreachable, and a format that distinguishes "key absent" from "key present and null" cannot be served by `dd` at all.

llm-gateway hits it on the OpenAI streaming contract, where every intermediate chunk must carry `"finish_reason": null` and the terminal chunk a value. the field is a `*string` with no `omitempty` precisely so `encoding/json` emits the null; there is a wire-shape test pinning it, because a client reading `finish_reason` as a presence check breaks when the key vanishes. that one requirement is enough to keep the package on `json` tags, which the `golang-conventions/df-binding` quality flags.

## discussion

verified against v1.0.5 — an intermediate chunk with a nil `finish_reason` tagged `+nullable`:

```
{"choices":[{"delta":{"content":"hi"},"index":0}],"id":"x","object":"chat.completion.chunk"}
```

`unbind.go` omits nil pointer fields before any tag is consulted, so the tag has no effect on output today. reading `+nullable` on both sides keeps one tag for one concept — this field's absence and its null mean the same thing, so the two directions agree — rather than adding a second write-only tag that could contradict it.

note for whoever takes it: `+nullable` is also needed on the read side far more widely than it looks. an untagged pointer field rejects an explicit null outright (`expected object for struct pointer, got <nil>`), which means every pointer in a struct modelling an external format needs the tag if the far side may send null. that is current behavior and not part of this card, but the two land together in practice.
