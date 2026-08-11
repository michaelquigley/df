---
title: secret values reach error strings
state: inbox
created: 2026-08-10
tags: [defect]
subsystems: [dd]
source: llm-gateway
---

values reach error strings at both decode stages, and `+secret` redacts at neither.

- **intake.** `DecodeStrictYAML` renders the raw lexeme: `$.keys[0].key: integer "0xdeadbeef" is not expressible as a JSON number`. reachable whenever an unquoted string is tagged `!!int` by YAML but is not a valid JSON number lexeme — hex, octal, leading zero.
- **binding.** `+secret` on a field changes nothing: `SecretRec.Key: expected string, got number 12345678` is byte-identical with and without the tag. `ConversionError` carries a `Value` field that renders into `Error()` on other paths too.

redact in both places. at binding, `+secret` is the signal and `TypeMismatchError`/`ConversionError` substitute a marker. at intake there is no struct to consult, so the value comes out unconditionally — the structural path and the reason carry all the diagnostic weight, and the lexeme adds nothing the path does not already locate.

## why

the intake half is the dangerous one: it runs *before* binding, so a consumer who installs a sanitizer around `dd.Bind` — the obvious place — never sees it. llm-gateway shipped exactly that gap in a first draft and caught it in review. each path individually needs a plaintext secret written as an unquoted non-string scalar, which for a hex- or digit-shaped key is what a YAML author produces without thinking. a refresh failure logs on every attempt, so one such record puts a live credential in the logs repeatedly, forever.

## discussion

surfaced while llm-gateway adopted `dd.Strict()` on a `v1.0.1` → `v1.0.2` bump, for virtual API key records written by third-party management planes. `key` is a legal plaintext field in that published contract, and the gateway's standing rule is that a secret never reaches a log line, unconditionally.

### intake

`DecodeStrictYAML` renders the offending scalar into its error:

```go
dd.DecodeStrictYAML([]byte("keys:\n  - name: a\n    key: 0xdeadbeef\n"))
→ $.keys[0].key: integer "0xdeadbeef" is not expressible as a JSON number

dd.DecodeStrictYAML([]byte("keys:\n  - name: a\n    key_sha256: 0x9f86d081\n"))
→ $.keys[0].key_sha256: integer "0x9f86d081" is not expressible as a JSON number

dd.DecodeStrictYAML([]byte("keys:\n  - name: a\n    key: 007\n"))
→ $.keys[0].key: integer "007" is not expressible as a JSON number
```

hex, octal, and leading-zero forms are all legal values in their key grammar, so a minted credential can look exactly like this. the structural path (`$.keys[0].key`) is the right thing to report and is already there; it is the quoted lexeme beside it that cannot be.

### binding

```go
type SecretRec struct {
    Name string `dd:"name"`
    Key  string `dd:"key,+secret"`
}
dd.BindYAML(s, []byte("name: a\nkey: 12345678\n"), dd.Strict())

→ binding field SecretRec.Key from key "key": SecretRec.Key: expected string, got number 12345678
```

byte-identical with and without the `+secret` tag.

### the workaround in the field

llm-gateway sanitizes the result of *every* decode stage, not just binding: any error whose structural path sits beneath a record is reduced to that path and `dd`'s text is dropped. it works, but it discards good diagnostic text everywhere to defend against two fields, and the path-not-stage scoping is a discipline every future consumer has to rediscover the hard way.
