# Opaque interface fields and nullable output work order

Implement [the combined design](opaque-interface-fields.md) in one stage. The behavior is concentrated in `dd/bind.go` and `dd/unbind.go`; the two cards should be verified together so opaque interface intake and nullable output agree.

## Existing integration points

`bindStruct` in `dd/bind.go` handles `+opaque` before normal conversion. It currently validates the field as exactly `map[string]any`, consumes its input key, applies required/nullable checks, asserts an object, and assigns the map directly. This is the extension point; changing `setNonPtrValue`'s general interface case would also change untagged behavior and is unnecessary.

`structToMap` in `dd/unbind.go` skips nil pointers before its `+omitempty` check and calls `valueToInterface` for ordinary output. `valueToInterface` omits nil interfaces, recursively unwraps non-nil ones, and already supports their concrete scalar, slice, and map values. Its conversions include custom marshalers, converters, `Dynamic`, and map-key collision detection. Nullable output belongs at the field boundary, before that conversion, rather than in the recursive value walker.

The tree has no `docs/current/` directory. Built behavior currently lives in `dd/README.md`, godoc, examples, and the documentation site under `docs/src/content/docs/`. Update those existing surfaces and add a compact `docs/current/` reference for the combined behavior when implementation lands.

## Stage 1

### Opaque binding

1. In the existing opaque branch, accept exactly `map[string]any` or a type whose kind is `reflect.Interface` and whose method count is zero. Keep declaration validation before input lookup, so a visited invalid declaration fails even if its key is absent. Preserve traversal of anonymous structs and skipped/unexported fields; do not introduce a separate whole-type validation pass.
2. Preserve name resolution, consumed-key bookkeeping, required checks, and the existing `raw == nil` nullable-as-absent path. Reject bare nil for an opaque interface lacking `+nullable` with a contextual `TypeMismatchError`. Do not recursively normalize typed nil values supplied through direct map binding; those remain concrete Go values, as they do for opaque maps today.
3. For a valid empty-interface field and non-nil raw value, assign the raw reflected value directly. Keep the map field's object assertion and assignment unchanged. Do not invoke converters, `Dynamic` binders, or unmarshalers during opaque assignment; the existing opaque branch already bypasses them. Preserve its current tag handling, including its existing `+match` behavior, rather than widening this change into tag precedence reform.
4. Update the `DdTag.Opaque` comment and tag grammar documentation in `dd/df.go`, `Bind` documentation, and `Strict` documentation. There is no parser change because the tag already exists.

### Nullable output

1. In `structToMap`, retain skip/extra handling and field naming, then apply the existing `+omitempty` check before deciding whether to emit a nullable null.
2. Add a small nil predicate for nullable fields: unwrap interfaces, stopping successfully at a nil interface; otherwise recognize a nil pointer, map, or slice. Do not dereference non-nil pointers or call `IsNil` on other kinds. Channels, functions, and unsafe pointers do not gain support through this predicate.
3. When `tag.Nullable` and the predicate matches, assign `out[name] = nil` and continue. Leave the existing nil-pointer omission for fields that did not take that path, and retain the existing value conversion and post-conversion empty-map omission for all other values. No changes to `valueToInterface` are needed.
4. Update the `Nullable` tag documentation and `Unbind` godoc. Correct its stale statement that ordinary interface values cannot be unbound. Do not claim `+opaque` bypasses output conversion.

### Verification

Add focused table-driven tests using the repository's assertion libraries. Assert key presence separately from value equality wherever null and omission differ.

- **Opaque shapes:** JSON string, boolean, number, list, object, and lists containing nested objects/nulls; forgiving and strict intake. Strict numbers remain `json.Number` with the authored spelling. Exercise YAML scalar/list/object input too. Untagged `any`, opaque interfaces with methods (including `Dynamic`), pointers to interfaces, and opaque typed slices still fail. Test a named empty interface and invalid declarations with missing input.
- **Assignment:** direct `Bind` preserves concrete scalar types and shares a map/slice with the input; mutate the source after binding to prove aliasing. `Merge` replaces a supplied opaque object wholesale, preserving existing data only when the field is absent or nullable-null. Keep existing opaque map object rejection covered.
- **Null intake and tags:** bare null fails without `+nullable`; with it, both binding modes preserve an existing field value. Missing or nullable-null with `+required` fails. Include a custom field name, an opaque field alongside `+extra`, and a nested struct. Cover a nullable opaque field inside a nil embedded pointer so input null neither allocates the parent nor becomes an unknown/extra key.
- **Strict boundary:** duplicate keys nested inside an opaque interface object or list must fail during intake; an unknown sibling must fail binding, while arbitrary keys within the opaque value succeed. Use an invalid YAML value inside the opaque subtree to verify it cannot bypass strict intake. Keep strict `Merge` rejection covered.
- **Nullable output:** nil pointers to primitive and struct types, nil interfaces, nil maps/slices, interfaces containing typed nil pointers/maps/slices, and non-nil values. Verify `+omitempty` precedence using the existing field-level emptiness semantics, including a non-nil interface holding a typed nil. Zero primitives must stay primitive values. Unsupported values must still return errors. Existing untagged nil output must not change.
- **Traversal and hooks:** named nested fields, flattened value/non-nil pointer embeds, nil embedded parents, skipped fields, and `+extra` collection behavior. A nil nullable pointer implementing `Marshaler` must emit null without invoking that method; use a method that fails if called. Non-nil converter/marshaler behavior must still execute, and an opaque interface containing a colliding-key map must still produce `KeyCollisionError` on output.
- **Consumer contract:** bind and unbind both stop-string and stop-list examples through JSON, plus a nullable opaque message. Verify streaming chunk output contains `finish_reason` with null while its pointer is nil and with a string when populated. Exercise raw `Unbind`, JSON, JSONL, and YAML output for explicit null, relying on their shared pipeline rather than duplicating every writer/file test.

Run targeted `go test ./dd -count=1`, then `make test` (the repository's Go tests and vet). Format changed Go files and run `git diff --check`. Build the documentation site if its source was changed and dependencies are available; report any unavailable check honestly. Run Terminus on the completed stage and handle findings one at a time with Michael before calling it complete.

### Documentation and handoff

Update `dd/README.md`, the dd guide at `docs/src/content/docs/guides/dd.md`, and relevant claims in `docs/src/content/docs/guides/df-framework.md`; sweep the strict example README and reference docs for now-stale object-only and nullable-intake-only claims. Include an interface example and the precedence of `+omitempty`. Add an `## Unreleased` changelog entry in the house format describing both behavior changes, with the nullable output migration explicit.

Preserve operator edits already present in both roadmap cards and `.ranger/order.yaml`; never edit the ordering file. Keep the cards in place under the repository's no-deletion rule. Before retiring this spec/work order after implementation, resolve any new links to them and preserve live deferred concerns. Record only rationale or unresolved state not recoverable from the final code and docs in the project journal. Leave all work uncommitted.
