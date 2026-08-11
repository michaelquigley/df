---
title: strict intake rejects yaml timestamps
state: inbox
created: 2026-08-10
tags: [defect, spike]
subsystems: [dd]
source: llm-gateway
---

strict YAML intake rejects the `!!timestamp` scalar tag, so a document marshalled by plain `yaml.v3` — which emits `expires_at: 2026-12-31T23:59:59Z` unquoted — is refused by `dd.Strict()`. df round-trips through itself fine because `dd.UnbindYAML` quotes timestamps; anyone using the default Go YAML marshaller does not.

preferred shape: admit `!!timestamp` at intake as a `time.Time` and let *binding* decide whether the target field accepts it. strict binding already refuses a `time.Time` arriving at a `string` field, so exactness is preserved where it is checkable, and intake stops adjudicating a question it lacks the type information to answer.

## why

the only one of the four llm-gateway requests with no consumer-side workaround. rejection happens at intake, before binding — the very boundary `DecodeStrictYAML` exists to let a consumer normalize across — so by the time a caller holds the tree the document is already refused. their remaining move is publishing "quote your timestamps" as a normalization rule on a third-party contract, which is the cross-implementation trap that section exists to remove.

## discussion

surfaced while llm-gateway adopted `dd.Strict()` on a `v1.0.1` → `v1.0.2` bump. it reads virtual API keys from external sources — a YAML file and an HTTP endpoint — with the record schema published as a contract that third-party management planes implement, so someone else's software writes the documents df has to decode.

that consumer position is what makes decoding strictness a security property there rather than a tidiness one, and it is the frame behind all four of the requests that came out of the adoption. a management plane that writes `allowed_model` instead of `allowed_models` produces a record whose restriction field is *absent*, absent means unrestricted, and a key intended for one model silently reaches every model — while the refresh succeeds and the gauges stay green.

the error as it lands:

```
$.keys[0].expires_at: scalar tag !!timestamp is rejected in strict mode — quote the value to bind it as a string
```

and the asymmetry that produces it:

```go
t := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
r := R{Name: "a", ExpiresAt: &t}

yaml.Marshal(r)   →  expires_at: 2026-12-31T23:59:59Z      // dd-strict REJECTS
dd.UnbindYAML(r)  →  expires_at: "2026-12-31T23:59:59Z"    // accepted
```

the message is about scalar tags rather than about quoting, so an author has to already know the rule to read it as advice.

three shapes were proposed, in preference order:

1. admit `!!timestamp` at intake as a `time.Time` value and let binding decide whether the target field accepts it.
2. admit `!!timestamp` only when it round-trips losslessly to RFC3339, rejecting the genuinely lossy YAML timestamp forms — bare dates like `2026-12-31`, space-separated forms — that are the real hazard.
3. keep the rejection, but make the error name the fix in terms an author will act on, and document the rule alongside `Strict()` so a contract publisher learns it before their consumers do.

the case for the first: the current rule reads like it is protecting exactness, but what it actually protects against is *YAML's timestamp grammar being wider than RFC3339* — a value question, not a syntax question, and the value question is answerable one layer down where the target type is known.

llm-gateway meanwhile carries "quote your timestamps" as a documented normalization rule on its published key-record contract, to be removed if this changes.
