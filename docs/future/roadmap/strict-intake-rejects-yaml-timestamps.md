---
title: strict intake rejects yaml timestamps
state: inbox
created: 2026-08-10
tags: [defect, spike]
subsystems: [dd]
source: docs/future/requests.md
---

strict YAML intake rejects the `!!timestamp` scalar tag, so a document marshalled by plain `yaml.v3` — which emits `expires_at: 2026-12-31T23:59:59Z` unquoted — is refused by `dd.Strict()`. df round-trips through itself fine because `dd.UnbindYAML` quotes timestamps; anyone using the default Go YAML marshaller does not.

preferred shape: admit `!!timestamp` at intake as a `time.Time` and let *binding* decide whether the target field accepts it. strict binding already refuses a `time.Time` arriving at a `string` field, so exactness is preserved where it is checkable, and intake stops adjudicating a question it lacks the type information to answer.

## why

the only one of the four llm-gateway requests with no consumer-side workaround. rejection happens at intake, before binding — the very boundary `DecodeStrictYAML` exists to let a consumer normalize across — so by the time a caller holds the tree the document is already refused. their remaining move is publishing "quote your timestamps" as a normalization rule on a third-party contract, which is the cross-implementation trap that section exists to remove.
