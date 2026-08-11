---
title: nullable tag
state: inbox
created: 2026-08-10
tags: [feature]
subsystems: [dd]
source: docs/future/requests.md
---

a `+nullable` struct tag, per field, where an explicit null binds as absent and leaves the zero value. today `allowed_models: null` fails with `expected array for slice, got <nil>`. it composes with `+required` the obvious way — a `+required +nullable` field given an explicit null still fails as required-missing — and reads naturally beside `+required` as the other half of the presence vocabulary.

a global `Options` flag would be the wrong shape: it makes nullability a property of the document rather than of the field, and contracts are not written that way.

## why

lowest priority of the four llm-gateway requests, and the one with a clean workaround: `DecodeStrictYAML` → strip the nullable keys → `dd.Bind` is documented as the seam for exactly this, is about fifteen lines, and is arguably more faithful than a library flag since the carve-out is named fields rather than a global posture. the case for building it anyway is that serializers emit `null` for empty collections and unset optionals as a matter of course, so every strict contract consuming machine-written documents meets this immediately and re-derives the same fifteen lines.
