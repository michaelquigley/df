# df Requests — from llm-gateway

Four requests against `github.com/michaelquigley/df/dd`, surfaced while building llm-gateway's dynamic key management against `dd.Strict()`. Ranked by value. **None of them block us** — every one has a workable gateway-side answer today, and the work proceeds against `v1.0.2` as shipped. They are collected here because three of the four are cheap, and one of them affects an external contract in a way no consumer-side workaround can reach.

## Context

llm-gateway is growing an extension point that reads virtual API keys from external sources — a YAML file and an HTTP endpoint — with the record schema published as a contract that third-party management planes implement. The gateway is the *consumer* of that contract; someone else's software writes the documents.

That makes decoding strictness a security property rather than a tidiness one. A management plane that writes `allowed_model` instead of `allowed_models` produces a record whose restriction field is *absent*, absent means unrestricted, and a key intended for one model silently reaches every model. The refresh succeeds and the gauges stay green. So: unknown fields, duplicate keys, and nulls where a value is required all have to reject the whole document.

`dd.Strict()` covers almost all of that out of the box, which is why llm-gateway is adopting it. Verified against the full rule set, on a `v1.0.1 → v1.0.2` bump with the gateway's suite green:

| rule | `dd.Strict()` |
|---|---|
| unknown fields, at envelope and nested-record level | yes |
| duplicate keys, YAML | yes |
| duplicate keys, JSON | yes |
| required fields missing | yes, via `+required` |
| null where a value is required | yes |
| no type coercion (`version: "1"`, `name: 42`) | yes |
| integer lexeme check (`count: 3.7`) | yes |
| multi-document YAML, JSON trailing data, YAML anchors | yes |
| errors name structure rather than values | mostly — two exceptions, both in request 3 |

The requests below are the residue.

## 1. `!!timestamp` at strict intake rejects what `yaml.v3` emits

**The behavior.** Strict YAML intake rejects the `!!timestamp` scalar tag:

```
$.keys[0].expires_at: scalar tag !!timestamp is rejected in strict mode — quote the value to bind it as a string
```

**Why it matters more than the others.** df round-trips through itself correctly — `dd.UnbindYAML` quotes timestamps, so a df writer feeding a df reader is fine. But plain `yaml.v3`, which is the default Go choice for anyone marshalling a document, does not:

```go
t := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
r := R{Name: "a", ExpiresAt: &t}

yaml.Marshal(r)   →  expires_at: 2026-12-31T23:59:59Z      // dd-strict REJECTS
dd.UnbindYAML(r)  →  expires_at: "2026-12-31T23:59:59Z"    // accepted
```

So a management plane that marshals its key file with `yaml.v3` produces a document the gateway refuses, and the resulting error is about scalar tags rather than about quoting — the author has to already know the rule to read the message as advice.

**And it is the one request with no consumer-side answer.** The other three can be worked around in llm-gateway. This one cannot, because the rejection happens at *intake*, before binding — which is precisely the boundary `DecodeStrictYAML` exists to let a consumer normalize across. By the time we hold the tree, the document has already been refused. The only gateway-side move left is documenting "quote your timestamps" as a normalization rule on a published contract, which is the kind of cross-implementation trap the contract's whole normalization section exists to remove.

**Suggested shapes**, in preference order:

1. Admit `!!timestamp` at intake as a `time.Time` value and let *binding* decide whether the target field accepts it. Strict binding already refuses a `time.Time` arriving at a `string` field, so the exactness guarantee is preserved where it is actually checkable; intake stops adjudicating a question it lacks the type information to answer.
2. Admit `!!timestamp` only when it round-trips losslessly to RFC3339, rejecting the genuinely lossy YAML timestamp forms (bare dates like `2026-12-31`, space-separated forms) that are the real hazard.
3. Keep the rejection but make the error name the fix in terms an author will act on, and document the rule alongside `Strict()` so a contract publisher learns it before their consumers do.

Option 1 is the one we'd pick. The current rule reads like it is protecting exactness, but what it actually protects against is *YAML's timestamp grammar being wider than RFC3339* — a value question, not a syntax question, and the value question is answerable one layer down where the target type is known.

## 2. `dd.Dynamic` and `dd.Strict()` do not compose

**The behavior.** The `type` discriminator that `dd` itself requires for a `Dynamic` field is then rejected by `dd`'s own strict binder as an unknown field:

```go
opts := dd.Strict()
opts.DynamicBinders = map[string]func(map[string]any) (dd.Dynamic, error){
    "file": func(m map[string]any) (dd.Dynamic, error) {
        v := &SrcCfg{}                      // Name, Path — no Type field
        if err := dd.Bind(v, m, dd.Strict()); err != nil {
            return nil, err
        }
        return v, nil
    },
}
dd.BindYAML(h, []byte("sources:\n  - type: file\n    name: local\n    path: /x\n"), opts)

→ Holder.Sources[0]: binding Dynamic type "file" failed: SrcCfg: unknown field "type"
```

The `Strict()` doc comment says Dynamic binders "remain in effect under strict mode," so this reads as a bug rather than a boundary. A binder can `delete(m, "type")` before binding, or the target can carry a throwaway `Type string` field, but both are workarounds for `dd` rejecting a key `dd` mandated.

**Suggested shape.** Have the strict binder treat the discriminator key as consumed for the struct a `DynamicBinder` produces — the same way `+extra` opts a struct out. `dd` put the key there; it should account for it.

**Not blocking us.** llm-gateway's heterogeneous `sources` list decodes under the forgiving posture (it is operator-authored configuration, not an external contract), so we never hit this in anger. It will bite the first person who tries to decode a discriminated union from a signed payload.

## 3. Secret values reach error strings at both decode stages

Two halves of one request: values reach error strings at **both** the intake and the binding stage, and `+secret` redacts at neither. They are filed together because a consumer who fixes only the half they have seen is left holding the other — which is exactly what happened to us.

### 3a. Strict intake quotes the offending scalar

`DecodeStrictYAML` renders the raw lexeme into its error:

```go
dd.DecodeStrictYAML([]byte("keys:\n  - name: a\n    key: 0xdeadbeef\n"))
→ $.keys[0].key: integer "0xdeadbeef" is not expressible as a JSON number

dd.DecodeStrictYAML([]byte("keys:\n  - name: a\n    key_sha256: 0x9f86d081\n"))
→ $.keys[0].key_sha256: integer "0x9f86d081" is not expressible as a JSON number

dd.DecodeStrictYAML([]byte("keys:\n  - name: a\n    key: 007\n"))
→ $.keys[0].key: integer "007" is not expressible as a JSON number
```

Reachable whenever an unquoted string value is tagged `!!int` by YAML but is not a valid JSON number lexeme — hex, octal, or a leading zero. Every one of those shapes is a legal value in our key grammar, so a minted credential can look exactly like this.

The structural path (`$.keys[0].key`) is the right thing to report and is already there. It is the quoted lexeme beside it that cannot be.

**This is the more dangerous half**, because intake runs *before* binding. A consumer who installs a sanitizer around `dd.Bind` — the obvious place, and where such a leak is first demonstrated — never sees these at all. We shipped precisely that gap in a first draft and caught it only in review.

### 3b. `+secret` does not redact values in strict binding errors

```go
type SecretRec struct {
    Name string `dd:"name"`
    Key  string `dd:"key,+secret"`
}
dd.BindYAML(s, []byte("name: a\nkey: 12345678\n"), dd.Strict())

→ binding field SecretRec.Key from key "key": SecretRec.Key: expected string, got number 12345678
```

Identical output with and without the `+secret` tag. `ConversionError` also carries a `Value` field that renders into `Error()`, so the same shape reaches other paths.

**Why both matter.** Each path is individually narrow — each needs a plaintext secret written as an unquoted non-string scalar, which for a hex- or digit-shaped key is exactly what a YAML author produces without thinking about it. But llm-gateway's rule is that a secret never reaches a log line, unconditionally, and `key` is a legal plaintext field in the published contract. A refresh failure gets logged on every attempt, so one such record puts a live credential in the logs repeatedly, forever.

**Suggested shape.** Redact the value in both places. At binding, `+secret` on the field is the signal, and `TypeMismatchError`/`ConversionError` substitute a marker. At intake there is no struct to consult, so the value has to come out unconditionally — the structural path and the reason ("integer is not expressible as a JSON number") carry all the diagnostic weight, and the lexeme adds nothing a path does not already locate.

**Our workaround.** llm-gateway sanitizes the result of *every* decode stage, not just binding: any error whose structural path sits beneath a record is reduced to that path and `dd`'s text is dropped. It works, but it discards good diagnostic text everywhere to defend against two fields, and the path-not-stage scoping is a discipline every future consumer has to rediscover the hard way.

## 4. A `+nullable` tag

**The behavior.** An explicit null at an optional field is an error rather than an absence:

```
allowed_models: null  →  expected array for slice, got <nil>
expires_at: null      →  expected time (RFC3339 string), got <nil>
```

**Why we want it.** Serializers emit `null` for an empty collection and an unset optional as a matter of course, so any strict contract consuming machine-written documents meets this immediately. In our contract three fields are explicitly nullable — a null `allowed_models` or `allowed_routes` means "no restriction," and `expires_at` is nullable outright — while `name`, `key`, `version`, and `keys` must keep rejecting a null.

**This one has a clean workaround and is genuinely lowest priority**, because `DecodeStrictYAML`/`DecodeStrictJSON` being public is documented as the seam for exactly this:

```go
m, err := dd.DecodeStrictYAML(data)   // syntax rules: dup keys, aliases, multi-doc
stripNulls(m, nullableFields)         // delete the three nullable keys when nil
err = dd.Bind(&doc, m, dd.Strict())   // binding rules: unknown fields, zero coercion
```

Verified working. About fifteen lines, and arguably more faithful than a library flag would be, since the carve-out is three named fields rather than a global posture.

**Suggested shape if you want it anyway.** A `+nullable` tag, per field, where an explicit null binds as absent and leaves the zero value. It composes with `+required` in the obvious way — a `+required +nullable` field given an explicit null still fails as required-missing — and it reads naturally beside `+required` as the other half of the presence vocabulary. A global `Options` flag would be the wrong shape: it would make nullability a property of the document rather than of the field, which is not how any of these contracts are written.

## What we are doing meanwhile

llm-gateway adopts `dd.Strict()` on the `v1.0.2` bump and ships with the workarounds above: the normalize step for request 4, path-scoped error sanitization across every decode stage for request 3, forgiving decoding of the `sources` list for request 2. Request 1 becomes a documented "quote your timestamps" rule on the published key-record contract, to be removed if df changes the intake behavior.

Nothing here is on our critical path. If a `v1.0.3` lands with request 1 addressed, the contract documentation gets simpler and one class of third-party integration failure disappears; the rest is ergonomics we can carry.
