---
title: secret values reach error strings
state: inbox
created: 2026-08-10
tags: [defect]
subsystems: [dd]
source: docs/future/requests.md
---

values reach error strings at both decode stages, and `+secret` redacts at neither.

- **intake.** `DecodeStrictYAML` renders the raw lexeme: `$.keys[0].key: integer "0xdeadbeef" is not expressible as a JSON number`. reachable whenever an unquoted string is tagged `!!int` by YAML but is not a valid JSON number lexeme — hex, octal, leading zero.
- **binding.** `+secret` on a field changes nothing: `SecretRec.Key: expected string, got number 12345678` is byte-identical with and without the tag. `ConversionError` carries a `Value` field that renders into `Error()` on other paths too.

redact in both places. at binding, `+secret` is the signal and `TypeMismatchError`/`ConversionError` substitute a marker. at intake there is no struct to consult, so the value comes out unconditionally — the structural path and the reason carry all the diagnostic weight, and the lexeme adds nothing the path does not already locate.

## why

the intake half is the dangerous one: it runs *before* binding, so a consumer who installs a sanitizer around `dd.Bind` — the obvious place — never sees it. llm-gateway shipped exactly that gap in a first draft and caught it in review. each path individually needs a plaintext secret written as an unquoted non-string scalar, which for a hex- or digit-shaped key is what a YAML author produces without thinking. a refresh failure logs on every attempt, so one such record puts a live credential in the logs repeatedly, forever.
